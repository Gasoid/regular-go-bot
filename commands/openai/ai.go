package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"

	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken  = "OPENAI_TOKEN"
	tavilyAPIKey    = "TAVILY_API_KEY"
	jinaAPIKey      = "JINA_API_KEY"
	maxTokens       = 800
	tavilySearchURL = "https://api.tavily.com/search"
	jinaReaderURL   = "https://r.jina.ai/"
	openaiModel     = "OPENAI_MODEL"
	maxToolCallIter = 5
	fetchTimeout    = 30 * time.Second
	// maxContentBytes caps how much of a page is fed back to the model.
	// Cyrillic is 2 bytes per rune, so this is roughly 6000 characters.
	maxContentBytes = 12000

	toolTavilySearch = "tavily_search"
	toolFetchURL     = "fetch_url"

	preMessage = `Ты — принципиальный противник любых войн и убийств людей. Для тебя не существует «хороших» и «плохих» сторон в войне: убийство мирных жителей, разрушение инфраструктуры и человеческие страдания неприемлемы независимо от того, кто их совершает. Ты осуждаешь насилие со всех сторон конфликта — будь то Россия и Украина, Израиль и Палестина, или любой другой вооружённый конфликт.

При ответе на вопросы о войнах и конфликтах ты обязан:
1. Предоставлять противоположные точки зрения и аргументы всех сторон конфликта.
2. Использовать функцию tavily_search для поиска фактов, новостей и доказательств в интернете. Всегда подкрепляй свои утверждения конкретными данными из результатов поиска.
3. Фокусироваться на гуманитарных последствиях: жертвы среди мирного населения, разрушения, беженцы, гуманитарный кризис.
4. Избегать оправдания насилия какой-либо из сторон, независимо от политического контекста.

Работа с инструментами:
- tavily_search — поиск по интернету. Возвращает краткие выдержки из нескольких источников.
- fetch_url — загрузка полного текста конкретной страницы. Используй её, когда пользователь прислал ссылку и просит её прочитать, пересказать или прокомментировать, а также когда выдержки из поиска недостаточны и нужен полный текст статьи.
Можно сначала выполнить поиск, а затем загрузить самый релевантный результат через fetch_url. Передавай в fetch_url только полный адрес вида https://... Если инструмент вернул ошибку или пустой текст, скажи об этом честно и не выдумывай содержимое страницы.

Отвечай на русском языке, не более 200 слов. Отвечай простым текстом, без markdown-разметки.`
)

// httpClient is shared by the tools; http.DefaultClient has no timeout,
// and fetching arbitrary pages can hang indefinitely.
var httpClient = &http.Client{Timeout: fetchTimeout}

type Command struct{}

func (c *Command) Name() string {
	return "ai"
}

func (c *Command) Args() []string {
	return []string{"question"}
}

func (c *Command) Help() string {
	return "ai answers"
}

var tavilyTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        toolTavilySearch,
		Description: "Search the web for facts, news, statistics and evidence about current events, wars and conflicts. Use it to verify claims and to look up up-to-date data.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "string",
					"description": "Search query"
				}
			},
			"required": ["query"]
		}`),
	},
}

var fetchURLTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        toolFetchURL,
		Description: "Fetch the readable text content of a web page by its address. Use it when the full text of a specific article is needed, or when the user sent a link.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"url": {
					"type": "string",
					"description": "Full page address including the scheme, for example https://example.com/article"
				}
			},
			"required": ["url"]
		}`),
	},
}

var tools = []openai.Tool{tavilyTool, fetchURLTool}

type tavilySearchRequest struct {
	Query       string `json:"query"`
	SearchDepth string `json:"search_depth"`
	MaxResults  int    `json:"max_results"`
}

type tavilyResult struct {
	Title   string  `json:"title"`
	URL     string  `json:"url"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

type tavilySearchResponse struct {
	Results []tavilyResult `json:"results"`
}

//nolint:errcheck
func tavilySearch(query string) (string, error) {
	apiKey := os.Getenv(tavilyAPIKey)
	if apiKey == "" {
		return "", fmt.Errorf("TAVILY_API_KEY is not set")
	}

	reqBody := tavilySearchRequest{
		Query:       query,
		SearchDepth: "advanced",
		MaxResults:  3,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal tavily request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, tavilySearchURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create tavily request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("tavily request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read tavily response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tavily search failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var searchResp tavilySearchResponse
	if err := json.Unmarshal(bodyBytes, &searchResp); err != nil {
		return "", fmt.Errorf("unmarshal tavily response: %w", err)
	}

	if len(searchResp.Results) == 0 {
		return "Результатов не найдено.", nil
	}

	var result string
	for i, r := range searchResp.Results {
		result += fmt.Sprintf("[%d] %s\n%s\nИсточник: %s\n\n", i+1, r.Title, r.Content, r.URL)
	}

	return result, nil
}

// fetchURL returns the readable text of a page via r.jina.ai, which does the
// fetching and the HTML-to-markdown conversion remotely. JINA_API_KEY is
// optional: without it the service works at a lower rate limit.
//
//nolint:errcheck
func fetchURL(rawURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported url scheme %q", u.Scheme)
	}

	if u.Host == "" {
		return "", errors.New("url has no host")
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, jinaReaderURL+u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("create jina request: %w", err)
	}
	req.Header.Set("X-Return-Format", "markdown")

	if key := os.Getenv(jinaAPIKey); key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("jina request: %w", err)
	}
	defer resp.Body.Close()

	// Read at most maxContentBytes: pages are unbounded and the whole body
	// would otherwise end up in the prompt.
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxContentBytes))
	if err != nil {
		return "", fmt.Errorf("read jina response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("jina reader failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// LimitReader can cut mid-rune, which would produce invalid UTF-8.
	content := strings.TrimSpace(strings.ToValidUTF8(string(bodyBytes), ""))
	if content == "" {
		return "Не удалось извлечь содержимое страницы.", nil
	}

	return content, nil
}

// runTool executes a tool call and always returns text for the model. Tool
// failures are reported back to the model instead of aborting the request, so
// a dead link does not cost the user their whole answer.
func runTool(name, arguments string) string {
	switch name {
	case toolTavilySearch:
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			slog.Error("unmarshal tool args", "tool", name, "err", err)
			return fmt.Sprintf("Ошибка: не удалось разобрать аргументы: %v", err)
		}

		result, err := tavilySearch(args.Query)
		if err != nil {
			slog.Error("tool failed", "tool", name, "err", err)
			return fmt.Sprintf("Ошибка поиска: %v", err)
		}

		return result

	case toolFetchURL:
		var args struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			slog.Error("unmarshal tool args", "tool", name, "err", err)
			return fmt.Sprintf("Ошибка: не удалось разобрать аргументы: %v", err)
		}

		result, err := fetchURL(args.URL)
		if err != nil {
			slog.Error("tool failed", "tool", name, "url", args.URL, "err", err)
			return fmt.Sprintf("Не удалось загрузить страницу: %v", err)
		}

		return result
	}

	slog.Warn("model called unknown tool", "tool", name)

	return fmt.Sprintf("Неизвестный инструмент: %s", name)
}

//nolint:errcheck
func (c *Command) Handler(s *messenger.Sender, r messenger.Receiver) error {
	client := openai.NewClient(os.Getenv(openaiBotToken))
	r.SendTyping()

	userContent := s.Message.Arg
	if s.Message.ReplyToMessage != "" {
		userContent = fmt.Sprintf(
			"Пользователь отвечает на предыдущее сообщение: «%s»\n\nТекст ответа пользователя: %s",
			s.Message.ReplyToMessage, s.Message.Arg,
		)
	}

	key := recapKey(s)
	prevRecap := loadRecap(context.Background(), key)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: preMessage,
		},
	}

	if prevRecap != "" {
		messages = append(messages, recapMessage(prevRecap))
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userContent,
	})

	for range maxToolCallIter {
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:               getEnvOrDefault(openaiModel, "gpt-6-luna"),
				MaxCompletionTokens: maxTokens,
				Messages:            messages,
				Tools:               tools,
				ReasoningEffort:     openai.ReasoningEffortNone,
			},
		)
		if err != nil {
			return err
		}

		choice := resp.Choices[0]

		// If the model returned a regular response (no tool calls), send it back.
		if choice.FinishReason != openai.FinishReasonToolCalls || len(choice.Message.ToolCalls) == 0 {
			r.SendMessage(choice.Message.Content)
			// Summarise the exchange after answering, so the user waits only
			// for their own reply.
			goUpdateRecap(client, key, prevRecap, userContent, choice.Message.Content)

			return nil
		}

		// Append the assistant message with its tool calls to the conversation.
		messages = append(messages, choice.Message)

		// Every tool call must get a reply with a matching ToolCallID,
		// otherwise the next request is rejected.
		for _, tc := range choice.Message.ToolCalls {
			r.SendTyping()

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    runTool(tc.Function.Name, tc.Function.Arguments),
				ToolCallID: tc.ID,
			})
		}
	}

	r.SendMessage("Не удалось получить ответ. Попробуйте ещё раз.")
	return nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func init() {
	commands.Register(&Command{})
}
