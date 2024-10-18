package telegram

import (
	"context"
	"os"
	"os/signal"

	"log/slog"

	"github.com/Gasoid/regular-go-bot/commands"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	telegramBotToken = "BOT_TOKEN"
	downloadFileUrl  = "https://api.telegram.org/file/bot%s/%s"
)

func Run() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{
		bot.WithDefaultHandler(defaultHandler),
	}

	b, err := bot.New(os.Getenv(telegramBotToken), opts...)
	if err != nil {
		slog.Error("bot.New", "err", err)
		os.Exit(1)
	}

	for _, c := range commands.List() {
		b.RegisterHandler(bot.HandlerTypeMessageText, "/"+c.Name(), bot.MatchTypePrefix, commandHandler(c))
	}

	commandList := []models.BotCommand{}
	for _, c := range commands.List() {
		commandList = append(commandList, models.BotCommand{Command: c.Name(), Description: c.Help()})
	}
	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: commandList, Scope: &models.BotCommandScopeAllPrivateChats{}})

	b.Start(ctx)
}
