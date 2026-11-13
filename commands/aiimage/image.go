package aiimage

import (
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/Gasoid/regular-go-bot/v3/messenger"

	"context"

	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken = "OPENAI_TOKEN"
)

type Command struct{}

func (c *Command) Name() string {
	return "image"
}

func (c *Command) Help() string {
	return "ai creates an image"
}

//nolint:errcheck
func (c *Command) Handler(s *messenger.Sender, r messenger.Receiver) error {
	ctx := context.Background()
	client := openai.NewClient(os.Getenv(openaiBotToken))
	reqUrl := openai.ImageRequest{
		Prompt:         s.Message.Text,
		Size:           openai.CreateImageSize1024x1024,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
		Model:          openai.CreateImageModelDallE3,
	}

	resp, err := client.CreateImage(ctx, reqUrl)
	if err != nil {
		return err
	}

	slog.Debug("image generated", "url", resp.Data[0].URL)
	httpResp, e := http.Get(resp.Data[0].URL)
	if e != nil {
		return e
	}
	defer httpResp.Body.Close()

	file, err := os.CreateTemp("", "image*.png")
	if err != nil {
		return err
	}

	defer os.Remove(file.Name())
	defer file.Close()

	_, err = io.Copy(file, httpResp.Body)
	if err != nil {
		return err
	}

	// r.SendPhoto(file.Name(), s.Message.Text)
	return nil
}

// func init() {
// 	commands.Register(&Command{})
// }
