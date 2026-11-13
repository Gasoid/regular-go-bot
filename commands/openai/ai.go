package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"

	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken  = "OPENAI_TOKEN"
	tavilyAPIKey    = "TAVILY_API_KEY"
	maxTokens       = 200
	tavilySearchURL = "https://api.tavily.com/search"
	maxToolCallIter = 3
	preMessage      = `Ты — принципиальный противник любых войн и убийств людей. Для тебя не существует «хороших» и «плохих» сторон в войне: убийство мирных жителей, разрушение инфраструктуры и человеческие страдания неприемлемы независимо от того, кто их совершает. Ты осуждаешь насилие со всех сторон конфликта — будь то Россия и Украина, Израиль и Палестина, или любой другой вооружённый конфликт.

При ответе на вопросы о войнах и конфликтах ты обязан:
1. Предоставлять противоположные точки зрения и аргументы всех сторон конфликта.
2. Использовать функцию tavily_search для поиска фактов, новостей и доказательств в интернете. Всегда подкрепляй свои утверждения конкретными данными из результатов поиска.
3. Фокусироваться на гуманитарных последствиях: жертвы среди мирного населения, разрушения, беженцы, гуманитарный кризис.
4. Избегать оправдания насилия какой-либо из сторон, независимо от политического контекста.

Отвечай на русском языке, не более 200 слов.`
)

type Command struct{}

func (c *Command) Name() string {
	return "ai"
}

func (c *Command) Help() string {
	return "ai answers"
}

var tavilyTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        "tavily_search",
		Description: "Поиск в интернете фактов, новостей, статистики и доказательств о текущих событиях, войнах и конфликтах. Используй для проверки утверждений и поиска актуальных данных.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"query": {
					"type": "string",
					"description": "Поисковый запрос для поиска информации"
				}
			},
			"required": ["query"]
		}`),
	},
}

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
		SearchDepth: "basic",
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

	resp, err := http.DefaultClient.Do(req)
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

//nolint:errcheck
func (c *Command) Handler(s *messenger.Sender, r messenger.Receiver) error {
	client := openai.NewClient(os.Getenv(openaiBotToken))

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: preMessage,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: s.Message.Arg,
		},
	}

	for range maxToolCallIter {
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:               openai.GPT4o,
				MaxCompletionTokens: maxTokens,
				Messages:            messages,
				Tools:               []openai.Tool{tavilyTool},
			},
		)
		if err != nil {
			return err
		}

		choice := resp.Choices[0]

		// If the model returned a regular response (no tool calls), send it back.
		if choice.FinishReason != openai.FinishReasonToolCalls || len(choice.Message.ToolCalls) == 0 {
			return r.SendMessage(choice.Message.Content)
		}

		// Append the assistant message with its tool calls to the conversation.
		messages = append(messages, choice.Message)

		// Execute each tool call and append results.
		for _, tc := range choice.Message.ToolCalls {
			if tc.Function.Name != "tavily_search" {
				continue
			}

			var args struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				return fmt.Errorf("unmarshal tool args: %w", err)
			}

			result, err := tavilySearch(args.Query)
			if err != nil {
				return err
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	return r.SendMessage("Не удалось получить ответ. Попробуйте ещё раз.")
}

func init() {
	commands.Register(&Command{})
}
