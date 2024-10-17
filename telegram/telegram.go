package telegram

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"log/slog"

	"github.com/Gasoid/regular-go-bot/commands"
	"github.com/Gasoid/regular-go-bot/metrics"
	"github.com/Gasoid/regular-go-bot/parsers"
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
		bot.WithDefaultHandler(defaultHandler),
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
				callback := commands.Callback{
					SendMessage: func(arg string) {
						b.SendMessage(ctx, &bot.SendMessageParams{
							ChatID: update.Message.Chat.ID,
							Text:   arg,
						})
					},
					SendVideo: func(filePath string) {
						f, err := os.Open(filePath)
						if err != nil {
							slog.Error("file not found", "err", err)
							return
						}
						defer f.Close()

						b.SendVideo(ctx, &bot.SendVideoParams{
							ChatID: update.Message.Chat.ID,
							Video: &models.InputFileUpload{
								Data:     f,
								Filename: "video",
							},
						})
					},
				}
				err := c.Handler(s, callback)
				if err != nil {
					slog.Error("handler failed", "err", err)
					return
				}

				metrics.CommandInc(c.Name())
			})
	}

	commandList := []models.BotCommand{}
	for _, c := range commands.List() {
		commandList = append(commandList, models.BotCommand{Command: c.Name(), Description: c.Help()})
	}
	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: commandList, Scope: &models.BotCommandScopeAllPrivateChats{}})

	b.Start(ctx)
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Voice != nil {
		f, err := b.GetFile(ctx, &bot.GetFileParams{
			FileID: update.Message.Voice.FileID,
		})
		if err != nil {
			slog.Error("b.GetFile", "err", err)
			return
		}

		for _, p := range parsers.ListVoiceParsers() {
			p.Handler(f.FilePath, parsers.Callback{
				ReplyMessage: func(text string) {
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   text,
						ReplyParameters: &models.ReplyParameters{
							MessageID: update.Message.ID,
							ChatID:    update.Message.Chat.ID,
						},
					})
				},
			})
		}
	}
}
