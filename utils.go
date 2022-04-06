package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	owm "github.com/briandowns/openweathermap"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	weatherIcons = map[int]string{
		2: "⚡️",
		3: "☔️",
		5: "🌧",
		6: "❄️",
		8: "🌤",
	}
)

func getLogs() string {
	resp, err := http.Get(os.Getenv(gistNewsURL))
	if err != nil {
		log.Printf("couldn't retrieve news %v", err)
		return "не могу получить новости"
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("couldn't retrieve news %v", err)
		return "не могу получить новости"
	}
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

func getWeather(cityName string) (*string, error) {
	var (
		icon string
		ok   bool
	)
	apiKey := os.Getenv(owmApiKey)
	w, err := owm.NewCurrent("C", "ru", apiKey)
	if err != nil {
		log.Println("couldn't load weather", err)
		return nil, fmt.Errorf("couldn't load weather %w", err)
	}
	w.CurrentByName(cityName)
	wDescr := ""
	for _, wW := range w.Weather {
		if wW.Description == "" {
			continue
		}
		if icon, ok = weatherIcons[wW.ID/100]; !ok {
			icon = defaultWeatherIcon
		}
		wDescr = fmt.Sprintf("%s%s%s ", wDescr, icon, wW.Description)
	}
	description := fmt.Sprintf(weatherTmpl, w.Name, wDescr, w.Main.Temp, w.Wind.Speed)
	return &description, nil
}
