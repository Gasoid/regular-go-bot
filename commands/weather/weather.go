package weather

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
	owm "github.com/briandowns/openweathermap"
)

const (
	owmApiKey     = "OWM_API_KEY"
	weatherCities = "DEFAULT_WEATHER_CITIES"
)

type Command struct{}

func (c *Command) Name() string {
	return "weather"
}

func (c *Command) Help() string {
	return "weather forecast, e.g.: /weather Los Angeles, US"
}

//nolint:errcheck
func (c *Command) Handler(s *messenger.Sender, r messenger.Receiver) error {
	var (
		text   string
		cities []string
		err    error
	)

	if s.Message.Arg != "" {
		cities = []string{s.Message.Arg}
	} else {
		cities, err = getDefaultCities()
		if err != nil {
			return err
		}
	}
	for _, city := range cities {
		description, err := c.getWeather(city, os.Getenv(owmApiKey))
		if err != nil {
			r.SendMessage("🧨 it doesn't look like a city name?!")
			return err
		}
		text = fmt.Sprintf("%s%s\n", text, description)
	}
	r.SendMessage(text)
	return nil
}

func getDefaultCities() ([]string, error) {
	cities := os.Getenv(weatherCities)
	if cities == "" {
		return nil, errors.New("no default cities")
	}

	return strings.Split(cities, ","), nil
}

func formatWeather(w *owm.CurrentWeatherData) string {
	var (
		icon string
		ok   bool
		// name, weather.description, main.temp, wind.speed
		weatherTmpl        = `📍 %s, %s🌡 %.1fC, 🌬 %.1fm/s`
		defaultWeatherIcon = "🌞"
		wDescr             = ""
		weatherIcons       = map[int]string{
			2: "⚡️",
			3: "☔️",
			5: "🌧",
			6: "❄️",
			8: "🌤",
		}
	)

	for _, wW := range w.Weather {
		if wW.Description == "" {
			continue
		}
		if icon, ok = weatherIcons[wW.ID/100]; !ok {
			icon = defaultWeatherIcon
		}
		wDescr = fmt.Sprintf("%s%s%s ", wDescr, icon, wW.Description)
	}

	return fmt.Sprintf(weatherTmpl, w.Name, wDescr, w.Main.Temp, w.Wind.Speed)
}

func (c *Command) getWeather(cityName string, owmApiKey string) (string, error) {
	w, err := owm.NewCurrent("C", "ru", owmApiKey)
	if err != nil {
		slog.Error("couldn't load weather", "err", err)
		return "", fmt.Errorf("couldn't load weather %w", err)
	}
	err = w.CurrentByName(cityName)
	if err != nil {
		return "", fmt.Errorf("couldn't load weather %w", err)
	}

	return formatWeather(w), nil
}

func init() {
	commands.Register(&Command{})
}
