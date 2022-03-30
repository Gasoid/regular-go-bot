package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"net/http"

	"github.com/Gasoid/workalendar/europe/germany/bavaria"
	"github.com/Gasoid/workalendar/europe/russia"
	"github.com/asvvvad/exchange"
)

func holiday(c *BotContext) {
	now := time.Now()

	if russia.IsHoliday(now) {
		h, _ := russia.GetHoliday(now)
		c.Text(fmt.Sprintf("Праздник сегодня: %s", h))
	} else {
		holidayBavaria, err := bavaria.GetHoliday(now)
		if err != nil {
			c.Text("Нету праздников сегодня")
		} else {
			c.Text(fmt.Sprintf("В РФ нету праздников, а в Баварии сегодня: %s", holidayBavaria))
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
	c.Text(fmt.Sprintf("**курс валют** \n$: %.2fруб \n€: %.2fруб \nBTC: %.2feur", usd, eur, btc))
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
	c.Text(fmt.Sprintf("Результаты отбора были объявлены 17 июня. Это было: %v", msgDuration))
}

func changelog(c *BotContext) {
	c.Text(getLogs())
}

func notFound(c *BotContext) {
	c.Text("Хм, не знаю такую команду")
}
