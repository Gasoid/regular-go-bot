package main

import (
	"fmt"
	"math/rand"
	"time"

	bot "github.com/Gasoid/regular-go-bot/bot"
)

func ozon(newMembersID map[int64]int) func(c *bot.BotContext) {
	return func(c *bot.BotContext) {
		if c.ChatTitle != "Gozone" {
			return
		}
		switch findKeyPhrase(c.Message) {
		case "github":
			c.Reply("🤔")
		case "play":
			c.Reply("😁")
		case "ozon":
			c.Reply("👿")
		}
		if c.NewChatMembers != nil {
			for _, member := range c.NewChatMembers {
				//rand.Seed(12000)
				answer := rand.Intn(200)
				newMembersID[member.ID] = answer + 1
				mention := fmt.Sprintf("@%s", member.Username)
				if member.Username == "" {
					mention = fmt.Sprintf("%s %s", member.FirstName, member.LastName)
				}
				c.Text("%s сколько будет 1 + %d = ? у тебя 2 минуты на ответ. Затем, пожалуйста, прочитай шапку группы!", mention, answer)
				c.AnswerFunc(func(b *bot.Bot) {
					ID := member.ID
					chatID := c.ChatID
					time.Sleep(2 * time.Minute)
					if _, ok := newMembersID[ID]; !ok {
						return
					}
					b.BanUser(chatID, ID)
					delete(newMembersID, ID)
				})
			}
		}

		if a, ok := newMembersID[c.User.ID]; ok {
			if c.Message == fmt.Sprint(a) {
				c.Reply("По-любому ты сделал задание E!?")
				delete(newMembersID, c.User.ID)
			}
		}
	}
}
