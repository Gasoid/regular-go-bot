package main

import (
	"time"
)

const (
	token       = "TOKEN"
	gistNewsURL = "GIST_LOGS_URL"
	endDate     = "Jun 17 2020"
	timeBreak   = 5 // seconds
	helpMessage = "набирай /changelog или /estimation.\nисходник: https://github.com/Gasoid/regular-go-bot"
	//typeCode    = "code"
	timeout    = 60
	enableOzon = "ENABLE_OZON"
	gozone     = "Gozone"
)

var lastUpdate time.Time

var phrases = map[string]string{
	"github":                  "github",
	"https://play.golang.org": "play",
	"https://ozon":            "ozon",
}

var newMembersID map[int64]int

func main() {
	bot := New()
	updates := bot.Updates()
	lastUpdate = time.Now()
	newMembersID = make(map[int64]int)
	bot.Command("help", help)
	bot.Command("estimation", estimation)
	bot.Command("changelog", changelog)
	bot.Command("currency", currency)
	bot.Command("joke", joke)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		bot.HandleOzon(&update)
		bot.HandleCommand(&update)
	}
}
