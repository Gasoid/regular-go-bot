package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	token       = "TOKEN"
	gistNewsURL = "GIST_LOGS_URL"
	endDate     = "Jun 17 2020"
)

func getLogs() string {
	resp, err := http.Get(os.Getenv(gistNewsURL))
	if err != nil {
		log.Printf("couldn't retrieve news %v", err)
		return "не смог получить новости"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv(token))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}

		// Create a new MessageConfig. We don't have text yet,
		// so we should leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = "набирай /changelog или /estimation."
		case "estimation":
			msgDuration := ""
			june17, _ := time.Parse("Jan 02 2006", endDate)
			duration := june17.Sub(time.Now())
			days := duration.Hours() / 24
			if days > 1 {
				msgDuration = fmt.Sprintf("%1.f дн", days)
			} else {
				msgDuration = fmt.Sprintf("%v часов", duration.Hours())
			}
			msg.Text = fmt.Sprintf("Результаты отбора будут объявлены 17 июня. Осталось: %v", msgDuration)
		case "changelog":
			msg.Text = getLogs()
		default:
			msg.Text = "Хм, не знаю такую команду"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}
