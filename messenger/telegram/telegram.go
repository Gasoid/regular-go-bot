package telegram

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func init() {
	messenger.RegisterService(new(tgInstance))
}

const (
	telegramBotToken = "TELEGRAM_BOT_TOKEN"
	downloadFileUrl  = "https://api.telegram.org/file/bot%s/%s"
)

type chat struct {
	sender   *messenger.Sender
	receiver messenger.Receiver
}

func (c *chat) GetSender() *messenger.Sender {
	return c.sender
}

func (c *chat) GetReceiver() messenger.Receiver {
	return c.receiver
}

type tgInstance struct{}

//nolint:errcheck
func (t *tgInstance) Run(ctx context.Context, chats chan<- messenger.Chat) error {
	token := os.Getenv(telegramBotToken)
	if token == "" {
		slog.Info("No telegram token provided, so telegram has not started")
		return nil
	}

	handler := func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}

		sender := &messenger.Sender{
			UserID:      strconv.FormatInt(update.Message.From.ID, 10),
			Username:    update.Message.From.Username,
			Group:       update.Message.Chat.Title,
			ChatID:      strconv.FormatInt(update.Message.Chat.ID, 10),
			MessageType: messenger.TextMessage,
			Message: messenger.Message{
				IsGroup: !update.Message.Chat.IsDirectMessages,
				Text:    update.Message.Text,
				Command: "",
			},
		}

		slog.Debug("is chat type a group", "check", update.Message.Chat.Type == models.ChatTypeGroup, "direct_message", update.Message.Chat.IsDirectMessages)

		switch {
		case strings.HasPrefix(update.Message.Text, "/"):
			parts := strings.Split(update.Message.Text, " ")
			sender.Message.Command, _ = strings.CutPrefix(parts[0], "/")
			sender.MessageType = messenger.Command
			sender.Message.Arg = strings.Join(parts[1:], " ")

		case update.Message.Voice != nil:
			f, err := b.GetFile(ctx, &bot.GetFileParams{
				FileID: update.Message.Voice.FileID,
			})
			if err != nil {
				slog.Error("b.GetFile", "err", err)
				return
			}

			sender.MessageType = messenger.VoiceMessage
			sender.Message.VoiceUrl = fmt.Sprintf(downloadFileUrl, os.Getenv(telegramBotToken), f.FilePath)
		case len(update.Message.Photo) > 0:
			sender.MessageType = messenger.Photo
			for _, p := range update.Message.Photo {
				f, err := b.GetFile(ctx, &bot.GetFileParams{FileID: p.FileID})
				if err != nil {
					slog.Error("b.GetFile", "err", err)
					continue
				}

				path, err := downloadFile(f.FilePath)
				if err != nil {
					slog.Error("can't download photo", "err", err)
					continue
				}
				sender.Message.PhotoPaths = append(sender.Message.PhotoPaths, path)
			}

		case update.Message.Location != nil:
			sender.MessageType = messenger.Location
			sender.Message.Coordinates = [2]float64{
				update.Message.Location.Latitude, update.Message.Location.Longitude,
			}
		case update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == b.ID():
			sender.MessageType = messenger.ReplyMessage
			sender.Message.ReplyToMessage = update.Message.ReplyToMessage.Text
			fallthrough
		case len(update.Message.Entities) > 0:
			u, err := b.GetMe(ctx)
			if err != nil {
				slog.Error("GetMe", "err", err)
				return
			}

			if strings.HasPrefix(update.Message.Text, "@"+u.Username) {
				sender.MessageType = messenger.Mentioned
			}
		}

		chats <- &chat{
			sender: sender,
			receiver: &messenger.Callback{
				SendMessageFunc: func(text string) {
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   text,
					})
				},
				ReplyMessageFunc: func(text string) {
					b.SendMessage(ctx, &bot.SendMessageParams{
						ChatID: update.Message.Chat.ID,
						Text:   text,
						ReplyParameters: &models.ReplyParameters{
							MessageID: update.Message.ID,
							ChatID:    update.Message.Chat.ID,
						},
					})
				},
			},
		}
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(token, opts...)
	if err != nil {
		return err
	}

	for _, c := range commands.List() {
		b.RegisterHandler(bot.HandlerTypeMessageText, "/"+c.Name(), bot.MatchTypePrefix, handler)
	}

	commandList := []models.BotCommand{}
	for _, c := range commands.List() {
		commandList = append(commandList, models.BotCommand{Command: c.Name(), Description: c.Help()})
	}
	b.SetMyCommands(ctx, &bot.SetMyCommandsParams{Commands: commandList, Scope: &models.BotCommandScopeAllPrivateChats{}})

	slog.Info("service is running")
	b.Start(ctx)
	return nil
}

//nolint:errcheck
func downloadFile(filepath string) (path string, err error) {
	suffix := strings.Split(filepath, ".")
	if len(suffix) < 2 {
		return "", errors.New("file name doesn't contain suffix")
	}

	file, err := os.CreateTemp("", "tg*."+suffix[1])
	if err != nil {
		return "", err
	}
	defer file.Close()

	resp, err := http.Get(fmt.Sprintf(downloadFileUrl, os.Getenv(telegramBotToken), filepath))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	slog.Debug("downloadFile", "filepath", file.Name())

	return file.Name(), nil
}
