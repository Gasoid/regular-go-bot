package youtube

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/Gasoid/regular-go-bot/commands"

	"github.com/kkdai/youtube/v2"
)

type Command struct{}

func (c *Command) Name() string {
	return "youtube"
}

func (c *Command) Help() string {
	return "the command imports youtube video into telegram"
}

func (c *Command) Handler(s string, callback commands.Callback) error {
	path, err := downloadVideo(s)
	if err != nil {
		slog.Error("downloadVideo", "err", err)
		return err
	}

	callback.SendVideo(path)
	defer os.Remove(path)
	return nil
}

func init() {
	commands.Register(&Command{})
}

func downloadVideo(yUrl string) (string, error) {

	// Создаем клиент YouTube
	client := youtube.Client{}

	// Получаем информацию о видео
	video, err := client.GetVideo(yUrl)
	if err != nil {
		slog.Error("Error getting video info", "err", err)
		return "", err
	}

	if time.Hour*2 < video.Duration {
		return "video is too long, supported up to 2 hours, 480p", nil
	}

	// Получаем форматы видео
	formats := video.Formats.WithAudioChannels() // Только форматы с аудио

	// Ищем формат с разрешением 480p
	var selectedFormat *youtube.Format
	for _, format := range formats {
		if format.Height == 480 {
			selectedFormat = &format
			break
		}
	}

	if selectedFormat == nil {
		return "", errors.New("no format found with 480p resolution")
	}

	// Открываем файл для записи
	file, err := os.CreateTemp("", "video_480p")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Загружаем видео в выбранном формате
	resp, _, err := client.GetStream(video, selectedFormat)
	if err != nil {
		return "", err
	}
	defer resp.Close()

	// Записываем содержимое видео в файл
	_, err = file.ReadFrom(resp)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}
