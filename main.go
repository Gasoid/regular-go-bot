package main

const (
	token       = "TOKEN"
	gistNewsURL = "GIST_LOGS_URL"
	endDate     = "Jun 17 2020"
	helpMessage = "/currency - курс валют \n/joke - шутка\n/holiday - какой сегодня праздник\nнабирай /changelog или /estimation.\nисходник: https://github.com/Gasoid/regular-go-bot"
	timeout     = 60
	enableOzon  = "ENABLE_OZON"
	gozone      = "Gozone"
)

var phrases = map[string]string{
	"github":                  "github",
	"https://play.golang.org": "play",
	"https://ozon":            "ozon",
}

var newMembersID map[int64]int

func main() {
	bot := New()
	updates := bot.Updates()
	newMembersID = make(map[int64]int)
	bot.Command("help", help)
	bot.Command("estimation", estimation)
	bot.Command("changelog", changelog)
	bot.Command("currency", currency)
	bot.Command("joke", joke)
	bot.Command("holiday", holiday)
	go runEndpoint()
	for update := range updates {
		if update.Message == nil {
			continue
		}
		bot.HandleOzon(&update)
		bot.HandleCommand(&update)
	}
}
