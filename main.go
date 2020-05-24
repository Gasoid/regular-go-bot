package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	token       = "TOKEN"
	gistNewsURL = "GIST_LOGS_URL"
	endDate     = "Jun 17 2020"
	timeBreak   = 30 // seconds
	helpMessage = "набирай /changelog или /estimation.\nисходник: https://github.com/Gasoid/regular-go-bot"
	typeCode    = "code"
)

var lastUpdate time.Time

func getLogs() string {
	resp, err := http.Get(os.Getenv(gistNewsURL))
	if err != nil {
		log.Printf("couldn't retrieve news %v", err)
		return "не могу получить новости"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}

func hasItBeen() bool {
	duration := time.Now().Sub(lastUpdate)
	return duration.Seconds() > timeBreak
}

func isCode(message *tgbotapi.Message) bool {
	return strings.Contains(message.Text, "github") && !strings.Contains(message.Text, "gasoid")
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
	lastUpdate = time.Now()
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}
		if !hasItBeen() {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.DisableWebPagePreview = true
		if isCode(update.Message) {
			msg.Text = "🤔"
			msg.ReplyToMessageID = update.Message.MessageID
			if _, err := bot.Send(msg); err != nil {
				log.Println(err.Error())
			}
			lastUpdate = time.Now()
			continue
		}
		if !update.Message.IsCommand() { // ignore any non-command Messages
			continue
		}
		lastUpdate = time.Now()
		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = helpMessage
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
			log.Println(err.Error())
		}
	}
}
