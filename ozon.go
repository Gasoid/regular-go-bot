package main

import (
	"fmt"
	"math/rand"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func checkOzon(c *BotContext) {
	switch findKeyPhrase(c.Update.Message) {
	case "github":
		c.Msg.Text = "🤔"
		c.Msg.ReplyToMessageID = c.Update.Message.MessageID
	case "play":
		c.Msg.Text = "😁"
		c.Msg.ReplyToMessageID = c.Update.Message.MessageID
	case "ozon":
		c.Msg.Text = "👿"
		c.Msg.ReplyToMessageID = c.Update.Message.MessageID
	}
	if c.Update.Message.NewChatMembers != nil {
		for _, member := range c.Update.Message.NewChatMembers {
			//rand.Seed(12000)
			answer := rand.Intn(200)
			newMembersID[member.ID] = answer + 1
			//c.Msg.ReplyToMessageID = c.Update.Message.MessageID
			mention := fmt.Sprintf("@%s", member.UserName)
			if member.UserName == "" {
				mention = fmt.Sprintf("%s %s", member.FirstName, member.LastName)
			}
			c.Msg.Text = fmt.Sprintf("%s сколько будет 1 + %d = ? у тебя 2 минуты на ответ. Затем, пожалуйста, прочитай шапку группы!", mention, answer)
			go func(ID int64, chatID int64) {
				time.Sleep(2 * time.Minute)
				if _, ok := newMembersID[ID]; !ok {
					return
				}
				conf := tgbotapi.BanChatMemberConfig{}
				conf.ChatID = chatID
				conf.UserID = ID
				c.Action = &conf
				//c.api.Send(conf)

				delete(newMembersID, ID)
			}(member.ID, c.Update.Message.Chat.ID)
		}
	}

	if a, ok := newMembersID[c.Update.Message.From.ID]; ok {
		if c.Update.Message.Text == fmt.Sprint(a) {
			c.Msg.Text = "По-любому ты сделал задание E!?"
			c.Msg.ReplyToMessageID = c.Update.Message.MessageID
			delete(newMembersID, c.Update.Message.From.ID)
		}
	}
}
