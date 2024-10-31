package location

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/Gasoid/regular-go-bot/parsers"

	"context"

	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken = "OPENAI_TOKEN"
)

type Command struct {
	name string
}

func (c *Command) Name() string {
	return "location"
}

func (c *Command) Handler(coords string, callback parsers.Callback) error {
	coordsList := strings.Split(coords, ",")
	if len(coordsList) < 2 {
		slog.Debug("location", "msg", "provided coordinates are wrong")
		return nil
	}

	lat, err := strconv.ParseFloat(coordsList[0], 64)
	if err != nil {
		return err
	}

	long, err := strconv.ParseFloat(coordsList[1], 64)
	if err != nil {
		return err
	}

	weatherText, err := c.getWeatherByCoords(lat, long)
	if err != nil {
		return err
	}

	callback.ReplyMessage(weatherText)

	client := openai.NewClient(os.Getenv(openaiBotToken))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:               openai.GPT4oMini,
			MaxCompletionTokens: 250,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "You provide interesting fact/history about location/place (radius 20Km) with no more than 150 words, should be in Russian language, please add appropriate emojis",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Location is %s with coordinates latitude: %f longitude: %f", c.name, lat, long),
				},
			},
		},
	)

	if err != nil {
		slog.Error("ChatCompletion error", "err", err)
		return nil
	}

	callback.ReplyMessage(resp.Choices[0].Message.Content)
	return nil
}

func init() {
	parsers.RegisterLocationParser(&Command{})
}
