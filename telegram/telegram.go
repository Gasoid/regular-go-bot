package telegram

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"log/slog"

	"github.com/Gasoid/regular-go-bot/commands"
	"github.com/Gasoid/regular-go-bot/metrics"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	telegramBotToken = "BOT_TOKEN"
)

func Run() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(helpHandler),
	}

	b, err := bot.New(os.Getenv(telegramBotToken), opts...)
	if err != nil {
		slog.Error("bot.New", "err", err)
		os.Exit(1)
	}

	for _, c := range commands.List() {
		commandPrefix := "/" + c.Name()
		b.RegisterHandler(bot.HandlerTypeMessageText, commandPrefix, bot.MatchTypePrefix,
			func(ctx context.Context, b *bot.Bot, update *models.Update) {
				s, _ := strings.CutPrefix(update.Message.Text, commandPrefix)

				callback := func(s string) {
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   s,
					})
				}

				err := c.Handler(s, callback)
				if err != nil {
					slog.Error("handler failed", "err", err)
					return
				}

				metrics.CommandInc(c.Name())
			})
	}
	b.Start(ctx)
}

func helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update == nil || update.Message == nil || update.Message.Text == "" {
		return
	}
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "help...",
	})
}
