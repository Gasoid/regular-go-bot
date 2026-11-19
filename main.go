package main

import (
	"log/slog"

	"github.com/Gasoid/regular-go-bot/v3/metrics"

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
	if err := metrics.Setup(); err != nil {
		slog.Error("couldn't setup and initialize metrics", "err", err)
	}

	kv.Init()
	start()
}
