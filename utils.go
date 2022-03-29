package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func findKeyPhrase(message *tgbotapi.Message) string {
	for k, v := range phrases {
		if strings.Contains(message.Text, k) && !strings.Contains(message.Text, "gasoid") {
			return v
		}
	}
	return ""
}
