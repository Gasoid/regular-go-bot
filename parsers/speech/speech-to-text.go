package speech

import (
	"log/slog"
	"os"

	"github.com/Gasoid/regular-go-bot/parsers"

	"context"

	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken = "OPENAI_TOKEN"
)

type Command struct{}

func (c *Command) Name() string {
	return "speech-to-voice"
}

func (c *Command) Handler(filepath string, callback parsers.Callback) error {
	slog.Debug("Speech", "msg", "started")

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

	slog.Debug("Speech", "msg", "finished")
	return nil
}

func init() {
	parsers.RegisterVoiceParser(&Command{})
}
