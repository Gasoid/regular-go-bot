package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Gasoid/regular-go-bot/v3/messenger"
)

const (
	maxMessages = 5
)

func start() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dispatcher := NewDispatcher()

	chats := make(chan messenger.Chat, maxMessages)

	for _, s := range messenger.ListServices() {
		go func() {
			err := s.Run(ctx, chats)
			if err != nil {
				slog.Error("can't run service", "service", s.Name(), "err", err)
			}
		}()
	}

	errChan := make(chan error, 1)
	go httpEndpoint(errChan)

	for {
		slog.Debug("waiting for incoming messages")
		select {
		case chat := <-chats:
			err := dispatcher.Process(chat.GetSender(), chat.GetReceiver())
			if err != nil {
				slog.Error("handler returns error", "error", err)
			}
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			healthSrv.Shutdown(shutdownCtx)
			slog.Info("Graceful shutdown")
			return
		case err := <-errChan:
			if !errors.Is(err, http.ErrServerClosed) {
				slog.Error("health endpoint returns error", "error", err)
			}
			return
		}
	}
}
