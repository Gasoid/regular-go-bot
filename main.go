package main

import (
	_ "github.com/Gasoid/regular-go-bot/v3/commands/help"
	_ "github.com/Gasoid/regular-go-bot/v3/commands/joke"
	_ "github.com/Gasoid/regular-go-bot/v3/commands/openai"
	_ "github.com/Gasoid/regular-go-bot/v3/commands/steam"
	_ "github.com/Gasoid/regular-go-bot/v3/commands/weather"
	"github.com/Gasoid/regular-go-bot/v3/kv"
	_ "github.com/Gasoid/regular-go-bot/v3/messenger/discord"
	_ "github.com/Gasoid/regular-go-bot/v3/messenger/telegram"
)

func main() {
	kv.Init()
	start()
}
