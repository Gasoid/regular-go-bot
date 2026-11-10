package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/Gasoid/regular-go-bot/v3/parsers/location"
	"github.com/Gasoid/regular-go-bot/v3/parsers/speech"
)

const (
	maxMessages = 5
)

func start() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dispatcher := NewDispatcher()

	chats := make(chan messenger.Chat, maxMessages)

	for _, s := range messenger.ListServices() {
		go func() {
			err := s.Run(ctx, chats)
			if err != nil {
				slog.Error("can't run service", "err", err)
			}
		}()
	}

	for {
		select {
		case chat := <-chats:
			err := dispatcher.Process(chat.GetSender(), chat.GetReceiver())
			if err != nil {
				slog.Error("handler returns error", "error", err)
			}
		case <-ctx.Done():
			cancel()
			slog.Info("Graceful shutdown")
			return
		}
	}
}

func voiceEventHandler(s *messenger.Sender, r messenger.Receiver) error {
	return speech.Handler(s, r)
}

func locationEventHandler(s *messenger.Sender, r messenger.Receiver) error {
	return location.Handler(s, r)
}
