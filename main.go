package main

import (
	"log/slog"

	"github.com/Gasoid/regular-go-bot/telegram"
)

func main() {
	httpEndpoint()
	slog.SetLogLoggerLevel(slog.LevelDebug)
	telegram.Run()
}
