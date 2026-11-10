package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/Gasoid/regular-go-bot/v3/parsers/location"
	"github.com/Gasoid/regular-go-bot/v3/parsers/speech"
)

const (
	maxMessages = 5
)

func start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	dispatcher := new(Dispatcher)

	dispatcher.eventFuncs[messenger.VoiceMessage] = voiceEventHandler
	dispatcher.eventFuncs[messenger.Location] = locationEventHandler

	for _, c := range commands.List() {
		dispatcher.commands[c.Name()] = c
	}

	dispatcher.middlewares = []func(messenger.Handler) messenger.Handler{
		metricsMiddleware,
	}

	chats := make(chan messenger.Chat, maxMessages)

	for _, s := range messenger.ListServices() {
		go s.Run(ctx, chats)
	}

	for {
		chat := <-chats

		err := dispatcher.Process(chat.GetSender(), chat.GetReceiver())
		if err != nil {
			slog.Error("handler returns error", "error", err)
		}

	}
}

func voiceEventHandler(s *messenger.Sender, r messenger.Receiver) error {
	return speech.Handler(s, r)
}

func locationEventHandler(s *messenger.Sender, r messenger.Receiver) error {
	return location.Handler(s, r)
}
