package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	owm "github.com/briandowns/openweathermap"
)

var (
	weatherIcons = map[int]string{
		2: "⚡️",
		3: "☔️",
		5: "🌧",
		6: "❄️",
		8: "🌤",
	}
	phrases = map[string]string{
		"github":                  "github",
		"https://play.golang.org": "play",
		"https://ozon":            "ozon",
	}
)

func getLogs(gistNewsURL string) string {
	resp, err := http.Get(gistNewsURL)
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

func findKeyPhrase(message string) string {
	for k, v := range phrases {
		if strings.Contains(message, k) && !strings.Contains(message, "gasoid") {
			return v
		}
	}
	return ""
}

func getWeather(cityName string, owmApiKey string) (*string, error) {
	var (
		icon string
		ok   bool
		// name, weather.description, main.temp, wind.speed
		weatherTmpl        = `📍 %s, %s🌡 %.1fC, 🌬 %.1fm/s`
		defaultWeatherIcon = "🌞"
	)
	w, err := owm.NewCurrent("C", "ru", owmApiKey)
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
