package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"net/http"

	b64 "encoding/base64"

	"github.com/Gasoid/workalendar/europe/germany/bavaria"
	"github.com/Gasoid/workalendar/europe/russia"
	"github.com/asvvvad/exchange"
	owm "github.com/briandowns/openweathermap"
	"github.com/goombaio/namegenerator"
	cbr "github.com/matperez/go-cbr-client"
	"github.com/thanhpk/randstr"
)

const (
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
	// name, weather.description, main.temp, wind.speed
	weatherTmpl = `📍 %s, %s🌡 %.1fC, 🌬 %.1fm/s`
	tokenLen    = 20
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

func encB64(c *BotContext) {
	arg := c.Update.Message.CommandArguments()
	if arg == "" {
		c.Text("🧨 no arguments, please send text")
		return
	}
	enc := b64.StdEncoding.EncodeToString([]byte(arg))
	c.Text("```%s```", enc)
}

func decB64(c *BotContext) {
	arg := c.Update.Message.CommandArguments()
	if arg == "" {
		c.Text("🧨 no arguments, please send base64 string")
		return
	}
	text, err := b64.StdEncoding.DecodeString(arg)
	if err != nil {
		c.Text("🧨 it is not base64 string")
		return
	}
	c.Text("```%s```", text)
}

func randomizer(c *BotContext) {
	var (
		len int
		err error
	)
	arg := c.Update.Message.CommandArguments()
	if arg != "" {
		len, err = strconv.Atoi(arg)
		if err != nil {
			log.Println("couldn't convert string to int:", err)
			len = tokenLen
		}
	} else {
		len = tokenLen
	}
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	nickName := nameGenerator.Generate()
	token := randstr.String(len)

	c.Text("🎲 Random phrase: %s\n🪄 Random nick: %s", token, nickName)
}

func weather(c *BotContext) {
	var (
		icon string
		ok   bool
		text string
	)
	apiKey := os.Getenv("OWM_API_KEY")
	w, err := owm.NewCurrent("C", "ru", apiKey)
	if err != nil {
		log.Println("couldn't load weather", err)
		return
	}
	cities := []string{"Saratov, RU", "Wuerzburg, DE", "Moscow, RU"}

	for _, city := range cities {
		w.CurrentByName(city)
		wDescr := ""
		for _, wW := range w.Weather {
			if wW.Description == "" {
				continue
			}
			if icon, ok = weatherIcons[wW.ID/100]; !ok {
				icon = "🌞"
			}
			wDescr = fmt.Sprintf("%s%s%s ", wDescr, icon, wW.Description)
		}
		description := fmt.Sprintf(weatherTmpl, w.Name, wDescr, w.Main.Temp, w.Wind.Speed)
		text = fmt.Sprintf("%s%s\n", text, description)
	}
	c.Text(text)
}

func chatInfo(c *BotContext) {
	c.Text("⚙️ ChatID: %d\nYour UserID: %d", c.Msg.ChatID, c.Update.Message.From.ID)
}

func holiday(c *BotContext) {
	now := time.Now()

	if russia.IsHoliday(now) {
		h, _ := russia.GetHoliday(now)
		c.Text("Праздник сегодня: %s", h)
	} else {
		holidayBavaria, err := bavaria.GetHoliday(now)
		if err != nil {
			c.Text("Нету праздников сегодня")
		} else {
			c.Text("В РФ нету праздников, а в Баварии сегодня: %s", holidayBavaria)
		}
	}
}

func currency(c *BotContext) {
	ex := exchange.New("USD")

	usd, err := ex.ConvertTo("RUB", 1)
	if err != nil {
		log.Print("exchange got error:", err.Error())
	}
	err = ex.SetBase("EUR")
	if err != nil {
		log.Print("exchange got error:", err.Error())
	}
	eur, err := ex.ConvertTo("RUB", 1)
	if err != nil {
		log.Print("exchange got error:", err.Error())
	}
	err = ex.SetBase("BTC")
	if err != nil {
		log.Print("exchange got error:", err.Error())
	}
	btc, err := ex.ConvertTo("EUR", 1)
	if err != nil {
		log.Print("exchange got error:", err.Error())
	}
	// err = ex.SetBase("ETH")
	// if err != nil {
	// 	log.Print("exchange got error:", err.Error())
	// }
	// eth, err := ex.ConvertTo("EUR", 1)
	// if err != nil {
	// 	log.Print("exchange got error:", err.Error())
	// }
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	cbrClient := cbr.NewClient()
	cbrUsd, err := cbrClient.GetRate("USD", now)
	if err != nil {
		log.Print("cbr got error:", err.Error())
	}
	yesterdayCbrUsd, err := cbrClient.GetRate("USD", yesterday)
	if err != nil {
		log.Print("cbr got error:", err.Error())
	}
	cbrUsdIcon := ""
	cbrEurIcon := ""
	if yesterdayCbrUsd < cbrUsd {
		cbrUsdIcon = "📈"
	}
	if yesterdayCbrUsd > cbrUsd {
		cbrUsdIcon = "📉"
	}

	cbrEur, err := cbrClient.GetRate("EUR", now)
	if err != nil {
		log.Print("cbr got error:", err.Error())
	}
	yesterdayCbrEur, err := cbrClient.GetRate("EUR", yesterday)
	if err != nil {
		log.Print("cbr got error:", err.Error())
	}
	if yesterdayCbrEur < cbrEur {
		cbrEurIcon = "📈"
	}
	if yesterdayCbrEur > cbrEur {
		cbrEurIcon = "📉"
	}
	c.Text(currencyMsgTmpl, cbrUsd, cbrUsdIcon, cbrEur, cbrEurIcon, usd, eur, btc)
}

func joke(c *BotContext) {
	reqURL := "https://jokesrv.rubedo.cloud/oneliner"
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Println("joke error:", err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("joke error:", err.Error())
		return
	}
	defer resp.Body.Close()
	rr := struct {
		Category string `json:"category"`
		Content  string `json:"content"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&rr); err != nil {
		log.Println("joke error:", err.Error())
		return

	}

	c.Text(rr.Content)
}

func help(c *BotContext) {
	c.Text(helpMessage)
}

func estimation(c *BotContext) {
	msgDuration := ""
	june17, _ := time.Parse("Jan 02 2006", endDate)
	if june17.Before(time.Now()) {
		c.Text("По идее результаты должны быть уже. Больше ничего не знаю!")
		return
	}

	duration := time.Until(june17)
	days := duration.Hours() / 24
	hours := duration.Hours() - float64(int(days)*24)
	if days > 1 {
		msgDuration = fmt.Sprintf("%1.f дн %1.f ч", days, hours)
	} else {
		msgDuration = fmt.Sprintf("%1.f часов", duration.Hours())
	}
	c.Text("Результаты отбора были объявлены 17 июня. Это было: %v", msgDuration)
}

func changelog(c *BotContext) {
	c.Text(getLogs())
}

func notFound(c *BotContext) {
	c.Text("Хм, не знаю такую команду")
}
