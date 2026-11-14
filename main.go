package main

import (
	_ "github.com/Gasoid/regular-go-bot/v3/commands/help"
	_ "github.com/Gasoid/regular-go-bot/v3/commands/joke"
	_ "github.com/Gasoid/regular-go-bot/v3/commands/openai"
	_ "github.com/Gasoid/regular-go-bot/v3/commands/weather"
	_ "github.com/Gasoid/regular-go-bot/v3/messenger/telegram"
)

func main() {
	start()
}
