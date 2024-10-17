package speech

import (
	"os"

	"github.com/Gasoid/regular-go-bot/parsers"

	"context"

	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken = "OPENAI_TOKEN"
)

type Command struct{}

func (c *Command) Handler(filepath string, callback parsers.Callback) error {
	client := openai.NewClient(os.Getenv(openaiBotToken))
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: filepath,
	}
	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		return err
	}

	callback.ReplyMessage(resp.Text)
	return nil
}

func init() {
	parsers.RegisterVoiceParser(&Command{})
}
