package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"net/http"

	b64 "encoding/base64"

	bot "github.com/Gasoid/regular-go-bot/bot"
	"github.com/Gasoid/workalendar/europe/germany/bavaria"
	"github.com/Gasoid/workalendar/europe/russia"
	"github.com/asvvvad/exchange"
	"github.com/goombaio/namegenerator"
	cbr "github.com/matperez/go-cbr-client"
	"github.com/thanhpk/randstr"
)

func timer(c *bot.BotContext) {
	if !c.HasArgs() {
		c.Text("🧨 please provide argument(minutes), e.g.: /timer 5")
		return
	}
	delay, err := strconv.Atoi(c.Arg())
	if err != nil {
		log.Println("couldn't convert arg to delay:", err)
		c.Text("🧨 couldn't convert your input to minute")
		return
	}
	c.Text("⏲ Set timer to %d minutes", delay)
	c.AnswerFunc(func(b *bot.Bot) {
		time.Sleep(time.Duration(delay) * time.Minute)
		c.Text("🔔 Timer (%d min) has finished", delay)
		b.Flush(c)
	})
}

func encB64(c *bot.BotContext) {
	if !c.HasArgs() {
		c.Text("🧨 no arguments, please send text")
		return
	}
	enc := b64.StdEncoding.EncodeToString([]byte(c.Arg()))
	c.Text("`%s`", enc)
}

func decB64(c *bot.BotContext) {
	if !c.HasArgs() {
		c.Text("🧨 no arguments, please send base64 string")
		return
	}
	text, err := b64.StdEncoding.DecodeString(strings.TrimSpace(c.Arg()))
	if err != nil {
		c.Text("🧨 it is not base64 string")
		return
	}
	c.Text("`%s`", string(text))
}

func randomizer(c *bot.BotContext) {
	var (
		len      int
		err      error
		tokenLen = 20
	)

	if c.Arg() != "" {
		len, err = strconv.Atoi(c.Arg())
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

func weather(apiKey string) func(c *bot.BotContext) {
	return func(c *bot.BotContext) {
		var (
			text string
		)
		cities := []string{"Saratov, RU", "Wuerzburg, DE", "Moscow, RU"}
		if c.Arg() != "" {
			cities = []string{c.Arg()}
		}
		for _, city := range cities {
			description, err := getWeather(city, apiKey)
			if err != nil {
				log.Println("couldn't get weather", err)
				c.Text("🧨 it doesn't look like a city name?!")
				return
			}
			text = fmt.Sprintf("%s%s\n", text, *description)
		}
		c.Text(text)
	}
}

func chatInfo(c *bot.BotContext) {
	c.Text("⚙️ ChatID: %d Your UserID: %d", c.ChatID, c.User.ID)
}

func holiday(c *bot.BotContext) {
	now := time.Now()

	if russia.IsHoliday(now) {
		h, _ := russia.GetHoliday(now)
		c.Text("Праздник сегодня: %s", h)
	} else {
		holidayBavaria, err := bavaria.GetHoliday(now)
		nextHoliday := russia.NextHoliday(now)
		b := bavaria.NextHoliday(now)
		if nextHoliday != nil && b != nil && b.Day.Sub(nextHoliday.Day).Hours() < 0 {
			nextHoliday = b
		} else {
			c.Text("Нету праздников сегодня")
			return
		}
		if err != nil {
			c.Text("Нету праздников сегодня. Следующий праздник: %s", nextHoliday.Name)
		} else {
			c.Text("В РФ нету праздников, а в Баварии сегодня: %s", holidayBavaria)
		}
	}
}

func currency(c *bot.BotContext) {
	currencyMsgTmpl := `
*КУРСЫ ВАЛЮТ*
🏛 ЦБ РФ:
$: %.2f руб %s
€: %.2f руб %s

FOREX:
$: %.2f руб
€: %.2f руб

🎲 CRYPTO:
BTC: %.2f eur
`

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

func joke(c *bot.BotContext) {
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

func estimation(c *bot.BotContext) {
	endDate := "Jun 17 2020"
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

func changelog(gistUrl string) func(c *bot.BotContext) {
	return func(c *bot.BotContext) {
		c.Text(getLogs(gistUrl))
	}
}

func help(helps *[]string) func(c *bot.BotContext) {
	helpMessage := "⛑ *Справка по командам* \n%s \n📃 исходник: https://github.com/Gasoid/regular-go-bot"
	return func(c *bot.BotContext) {
		text := strings.Join(*helps, "\n")
		c.Text(helpMessage, text)
	}
}
