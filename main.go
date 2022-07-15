package main

import (
	"log"
	"os"

	"github.com/Gasoid/regular-go-bot/bot"
	"github.com/Gasoid/regular-go-bot/metrics"
	"github.com/Gasoid/regular-go-bot/telegram"
)

const (
	token         = "TOKEN"
	gistNewsURL   = "GIST_LOGS_URL"
	weatherApiKey = "OWM_API_KEY"
)

func main() {
	helps := &[]string{}
	newMembersID := map[int64]int{}
	t, err := telegram.New(os.Getenv(token))
	if err != nil {
		log.Println("token is invalid", err.Error())
		return
	}
	httpEndpoint()
	bot := bot.New(t, helps, newMembersID)
	bot.Command("help", help(helps), "помощь по командам")
	bot.Command("estimation", estimation, "")
	bot.Command("changelog", changelog(os.Getenv(gistNewsURL)), "")
	bot.Command("currency", currency, "курс валют")
	bot.Command("joke", joke, "шутка")
	bot.Command("holiday", holiday, "какой сегодня праздник")
	bot.Command("weather", weather(os.Getenv(weatherApiKey)), "погода, например: /weather Los Angeles, US")
	bot.Command("chat_info", chatInfo, "chat info")
	bot.Command("random", randomizer, "randomizer")
	bot.Command("b64encode", encB64, "the command encodes string to base64")
	bot.Command("b64decode", decB64, "the command decodes base64 string")
	bot.Command("timer", timer, "will notify you in N minute, ex: /timer 10")
	bot.MessageHandler(ozon(newMembersID))
	bot.Run(metrics.New())
}
