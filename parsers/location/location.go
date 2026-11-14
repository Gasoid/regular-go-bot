package location

import (
	"fmt"
	"os"

	"context"

	"github.com/Gasoid/regular-go-bot/v3/messenger"
	openai "github.com/sashabaranov/go-openai"
)

const (
	openaiBotToken = "OPENAI_TOKEN"
	openaiModel    = "OPENAI_MODEL"
)

//nolint:errcheck
func Handler(sender *messenger.Sender, r messenger.Receiver) error {
	r.SendTyping()
	lat, long := sender.Message.Coordinates[0], sender.Message.Coordinates[1]

	city, weatherText, err := getWeatherByCoords(lat, long)
	if err != nil {
		return err
	}

	r.ReplyMessage(weatherText)

	client := openai.NewClient(os.Getenv(openaiBotToken))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:               model(),
			MaxCompletionTokens: 250,
			ReasoningEffort:     openai.ReasoningEffortNone,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "You provide interesting fact/history about location/place (radius 20Km) with no more than 150 words, should be in Russian language, please add appropriate emojis",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Location is %s with coordinates latitude: %f longitude: %f", city, lat, long),
				},
			},
		},
	)

	if err != nil {
		return err
	}

	r.ReplyMessage(resp.Choices[0].Message.Content)
	return nil
}

func model() string {
	if v := os.Getenv(openaiModel); v != "" {
		return v
	}
	return "gpt-6-luna"
}
