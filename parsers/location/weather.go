package location

import (
	"fmt"
	"log/slog"
	"os"

	owm "github.com/briandowns/openweathermap"
)

const (
	owmApiKey = "OWM_API_KEY"
)

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

func getWeatherByCoords(lat, long float64) (string, string, error) {
	w, err := owm.NewCurrent("C", "ru", os.Getenv(owmApiKey))
	if err != nil {
		slog.Error("couldn't load weather", "err", err)
		return "", "", fmt.Errorf("couldn't load weather %w", err)
	}

	err = w.CurrentByCoordinates(&owm.Coordinates{Longitude: long, Latitude: lat})
	if err != nil {
		return "", "", fmt.Errorf("couldn't load weather %w", err)
	}

	return w.Name, formatWeather(w), nil
}
