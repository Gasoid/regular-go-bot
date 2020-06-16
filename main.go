package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
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
	timeBreak   = 5 // seconds
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

func isOzonEmployee(firstName, secondName string) bool {
	return false
}

var phrases = map[string]string{
	"github":                  "github",
	"https://play.golang.org": "play",
	"https://ozon":            "ozon",
}

var newMembersID map[int]int

func findKeyPhrase(message *tgbotapi.Message) string {
	for k, v := range phrases {
		if strings.Contains(message.Text, k) && !strings.Contains(message.Text, "gasoid") {
			return v
		}
	}
	return ""
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
	newMembersID = make(map[int]int)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}
		// if !hasItBeen() {
		// 	continue
		// }
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		msg.DisableWebPagePreview = true
		switch findKeyPhrase(update.Message) {
		case "github":
			msg.Text = "🤔"
			msg.ReplyToMessageID = update.Message.MessageID
		case "play":
			msg.Text = "😁"
			msg.ReplyToMessageID = update.Message.MessageID
		case "ozon":
			msg.Text = "👿"
			msg.ReplyToMessageID = update.Message.MessageID
		}
		if update.Message.NewChatMembers != nil {
			for _, member := range *update.Message.NewChatMembers {
				//rand.Seed(12000)
				answer := rand.Intn(200)
				newMembersID[member.ID] = answer + 1
				//msg.ReplyToMessageID = update.Message.MessageID
				mention := fmt.Sprintf("@%s", member.UserName)
				if member.UserName == "" {
					mention = fmt.Sprintf("%s %s", member.FirstName, member.LastName)
				}
				msg.Text = fmt.Sprintf("%s сколько будет 1 + %d = ? у тебя 2 минуты на ответ. Затем, пожалуйста, прочитай шапку группы!", mention, answer)
				go func(ID int, chatID int64) {
					time.Sleep(2 * time.Minute)
					if _, ok := newMembersID[ID]; !ok {
						return
					}
					conf := tgbotapi.KickChatMemberConfig{}
					conf.ChatID = chatID
					conf.UserID = ID
					//conf.UntilDate =
					bot.KickChatMember(conf)
					delete(newMembersID, ID)
				}(member.ID, update.Message.Chat.ID)
			}
		}

		if a, ok := newMembersID[update.Message.From.ID]; ok {
			if update.Message.Text == fmt.Sprint(a) {
				msg.Text = "По-любому ты сделал задание E!?"
				msg.ReplyToMessageID = update.Message.MessageID
				delete(newMembersID, update.Message.From.ID)
			}
		}

		if update.Message.IsCommand() {
			// Extract the command from the Message.
			switch update.Message.Command() {
			case "help":
				msg.Text = helpMessage
			case "estimation":
				msgDuration := ""
				june17, _ := time.Parse("Jan 02 2006", endDate)
				if june17.Before(time.Now()) {
					msg.Text = fmt.Sprintf("По идее результаты должны быть уже. Больше ничего не знаю!")
					break
				}
				duration := june17.Sub(time.Now())
				days := duration.Hours() / 24
				hours := duration.Hours() - float64(int(days)*24)
				if days > 1 {
					msgDuration = fmt.Sprintf("%1.f дн %1.f ч", days, hours)
				} else {
					msgDuration = fmt.Sprintf("%1.f часов", duration.Hours())
				}
				msg.Text = fmt.Sprintf("Результаты отбора будут объявлены 17 июня. Осталось: %v", msgDuration)
			case "changelog":
				msg.Text = getLogs()
			default:
				msg.Text = "Хм, не знаю такую команду"
			}

		}

		if _, err := bot.Send(msg); err != nil {
			lastUpdate = time.Now()
			log.Println(err.Error())
		}
	}
}
