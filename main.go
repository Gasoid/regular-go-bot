package main

const (
	tokenLen           = 20
	token              = "TOKEN"
	gistNewsURL        = "GIST_LOGS_URL"
	endDate            = "Jun 17 2020"
	timeout            = 60
	enableOzon         = "ENABLE_OZON"
	gozone             = "Gozone"
	defaultWeatherIcon = "🌞"
	owmApiKey          = "OWM_API_KEY"
	helpMessage        = "⛑ ### Справка по командам ###\n%s \n📃 исходник: https://github.com/Gasoid/regular-go-bot"
	// name, weather.description, main.temp, wind.speed
	weatherTmpl     = `📍 %s, %s🌡 %.1fC, 🌬 %.1fm/s`
	currencyMsgTmpl = `
**КУРСЫ ВАЛЮТ**
🏛 ЦБ РФ:
$: %.2f руб %s
€: %.2f руб %s

FOREX:
$: %.2f руб
€: %.2f руб

🎲 CRYPTO:
BTC: %.2f eur
`
)

var (
	phrases = map[string]string{
		"github":                  "github",
		"https://play.golang.org": "play",
		"https://ozon":            "ozon",
	}

	newMembersID map[int64]int
	helps        = []string{}
)

func main() {
	bot := New()
	updates := bot.Updates()
	newMembersID = make(map[int64]int)
	bot.Command("help", help, "помощь по командам")
	bot.Command("estimation", estimation, "")
	bot.Command("changelog", changelog, "")
	bot.Command("currency", currency, "курс валют")
	bot.Command("joke", joke, "шутка")
	bot.Command("holiday", holiday, "какой сегодня праздник")
	bot.Command("weather", weather, "погода, например: /weather Los Angeles, US")
	bot.Command("chat_info", chatInfo, "chat info")
	bot.Command("random", randomizer, "randomizer")
	bot.Command("b64encode", encB64, "the command encodes string to base64")
	bot.Command("b64decode", decB64, "the command decodes base64 string")
	go runEndpoint()
	for update := range updates {
		if update.Message == nil {
			continue
		}
		bot.HandleOzon(&update)
		bot.HandleCommand(&update)
	}
}
