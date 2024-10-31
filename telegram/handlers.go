package telegram

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Gasoid/regular-go-bot/commands"
	"github.com/Gasoid/regular-go-bot/parsers"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const (
	downloadFileUrl = "https://api.telegram.org/file/bot%s/%s"
)

func commandHandler(c commands.Command) func(ctx context.Context, b *bot.Bot, update *models.Update) {
	commandPrefix := "/" + c.Name()
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
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
			SendPhoto: func(path, caption string) {
				fileData, err := os.ReadFile(path)
				if err != nil {
					slog.Error("file not found", "err", err)
					return
				}

				params := &bot.SendPhotoParams{
					ChatID:  update.Message.Chat.ID,
					Photo:   &models.InputFileUpload{Filename: "image.png", Data: bytes.NewReader(fileData)},
					Caption: caption,
				}

				b.SendPhoto(ctx, params)
			},
		}

		err := c.Handler(s, callback)
		if err != nil {
			slog.Error("handler failed", "err", err)
			return
		}
	}
}

func defaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message.Location != nil {
		for _, p := range parsers.ListLocationParsers() {
			err := p.Handler(fmt.Sprintf("%f,%f", update.Message.Location.Latitude, update.Message.Location.Longitude), parsers.Callback{
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
			if err != nil {
				slog.Error("p.Handler", "err", err)
			}
		}
	}

	if update.Message.Voice != nil {
		f, err := b.GetFile(ctx, &bot.GetFileParams{
			FileID: update.Message.Voice.FileID,
		})
		if err != nil {
			slog.Error("b.GetFile", "err", err)
			return
		}

		path, err := downloadVoice(f.FilePath)
		if err != nil {
			slog.Error("downloadVoice", "err", err)
			return
		}
		defer os.Remove(path)

		for _, p := range parsers.ListVoiceParsers() {
			err := p.Handler(path, parsers.Callback{
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
			if err != nil {
				slog.Error("p.Handler", "err", err)
			}
		}
	}
}

// func isChatPrivate(t models.ChatType) bool {
// 	return t == models.ChatTypePrivate
// }

func downloadVoice(filepath string) (path string, err error) {
	suffix := strings.Split(filepath, ".")
	if len(suffix) < 2 {
		return "", errors.New("file name doesn't contain suffix")
	}

	file, err := os.CreateTemp("", "voice*."+suffix[1])
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

	slog.Debug("downloadVOice", "filepath", file.Name())

	return file.Name(), nil
}
