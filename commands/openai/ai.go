package joke

import (
	"log/slog"
	"os"

	"github.com/Gasoid/regular-go-bot/commands"

	"context"

	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken = "OPENAI_TOKEN"
	maxTokens      = 100
	preMessage     = "You answer with no more than 100 words, should be in Russian language"
)

type Command struct{}

func (c *Command) Name() string {
	return "ai"
}

func (c *Command) Help() string {
	return "ai answers"
}

func (c *Command) Handler(message string, callback func(string)) error {
	client := openai.NewClient(os.Getenv(openaiBotToken))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:               openai.GPT4oMini,
			MaxCompletionTokens: maxTokens,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: preMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: message,
				},
			},
		},
	)

	if err != nil {
		slog.Error("ChatCompletion error", "err", err)
		return nil
	}

	callback(resp.Choices[0].Message.Content)
	return nil
}

func init() {
	commands.Register(&Command{})
}
