package speech

import (
	"os"

	"context"

	"github.com/Gasoid/regular-go-bot/v3/messenger"
	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken = "OPENAI_TOKEN"
)

func Handler(sender *messenger.Sender, r messenger.Receiver) error {
	client := openai.NewClient(os.Getenv(openaiBotToken))
	ctx := context.Background()

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: sender.Message.VoiceFilePath,
	}

	resp, err := client.CreateTranscription(ctx, req)
	if err != nil {
		return err
	}

	return r.ReplyMessage(resp.Text)
}
