package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	ai "github.com/Gasoid/regular-go-bot/v3/commands/openai"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/Gasoid/regular-go-bot/v3/parsers/location"
	"github.com/Gasoid/regular-go-bot/v3/parsers/speech"
)

type Dispatcher struct {
	commands    map[string]commands.Command
	middlewares []func(messenger.Handler) messenger.Handler
	eventFuncs  map[messenger.EventType]func(*messenger.Sender, messenger.Receiver) error
}

func (d *Dispatcher) Process(s *messenger.Sender, r messenger.Receiver) error {
	h := processMiddlewares(
		messenger.HandlerFunc(func(sender *messenger.Sender, receiver messenger.Receiver) error {
			switch sender.MessageType {
			case messenger.Command:
				if c, ok := d.commands[sender.Message.Command]; ok {
					return c.Handler(sender, receiver)
				}
			default:
				if f, ok := d.eventFuncs[sender.MessageType]; ok {
					return f(sender, receiver)
				}
			}

			return nil
		}), d.middlewares...)

	err := h.Process(s, r)
	if err != nil {
		return err
	}

	return nil
}

func voiceEventHandler(s *messenger.Sender, r messenger.Receiver) error {
	if s.Message.VoiceUrl != "" {
		filename := "voice.ogg"

		if _, r, found := strings.CutLast(s.Message.VoiceUrl, "/"); found && strings.Contains(r, ".") {
			filename = r
		}

		p, err := downloadFile(s.Message.VoiceUrl, filename)
		if err != nil {
			return err
		}

		s.Message.VoiceFilePath = p
	}

	//nolint:errcheck
	defer os.Remove(s.Message.VoiceFilePath)

	return speech.Handler(s, r)
}

func locationEventHandler(s *messenger.Sender, r messenger.Receiver) error {
	return location.Handler(s, r)
}

func messageEventHandler(s *messenger.Sender, r messenger.Receiver) error {
	const longMessageLength = 30
	if s.Message.IsGroup {
		words := strings.Split(s.Message.Text, " ")
		if len(words) >= longMessageLength || s.MessageType == messenger.Mentioned || s.MessageType == messenger.ReplyMessage {
			c := ai.Command{}
			s.Message.Arg = s.Message.Text
			return c.Handler(s, r)
		}
	}
	return nil
}

func photoEventsHandler(s *messenger.Sender, r messenger.Receiver) error {
	defer func() {
		for _, p := range s.Message.PhotoPaths {
			//nolint:errcheck
			os.Remove(p)
		}
	}()

	return nil
}

func NewDispatcher() *Dispatcher {
	dispatcher := new(Dispatcher)
	dispatcher.commands = make(map[string]commands.Command, len(commands.List()))
	dispatcher.eventFuncs = make(map[messenger.EventType]func(*messenger.Sender, messenger.Receiver) error)

	dispatcher.eventFuncs[messenger.VoiceMessage] = voiceEventHandler
	dispatcher.eventFuncs[messenger.Location] = locationEventHandler
	dispatcher.eventFuncs[messenger.TextMessage] = messageEventHandler
	dispatcher.eventFuncs[messenger.ReplyMessage] = messageEventHandler
	dispatcher.eventFuncs[messenger.Mentioned] = messageEventHandler
	dispatcher.eventFuncs[messenger.Photo] = photoEventsHandler

	for _, c := range commands.List() {
		dispatcher.commands[c.Name()] = c
	}

	dispatcher.middlewares = []func(messenger.Handler) messenger.Handler{
		metricsMiddleware,
	}

	return dispatcher
}

//nolint:errcheck
func downloadFile(fileUrl, filename string) (path string, err error) {
	suffix := strings.Split(filename, ".")
	if len(suffix) < 2 {
		return "", errors.New("file name doesn't contain suffix")
	}

	file, err := os.CreateTemp("", "voice*."+suffix[1])
	if err != nil {
		return "", err
	}
	defer file.Close()

	resp, err := http.Get(fileUrl)
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
