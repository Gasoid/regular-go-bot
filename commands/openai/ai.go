package openai

import (
	"os"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"

	"context"

	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken = "OPENAI_TOKEN"
	maxTokens      = 200
	preMessage     = "You answer with no more than 200 words, should be in Russian language, please add appropriate emojis"
)

type Command struct{}

func (c *Command) Name() string {
	return "ai"
}

func (c *Command) Help() string {
	return "ai answers"
}

func (c *Command) Handler(s *messenger.Sender, r messenger.Receiver) error {
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
					Content: s.Message.Arg,
				},
			},
		},
	)

	if err != nil {
		return err
	}

	r.SendMessage(resp.Choices[0].Message.Content)
	return nil
}

func init() {
	commands.Register(&Command{})
}
