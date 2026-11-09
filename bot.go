package main

import (
	"log/slog"

	"github.com/Gasoid/regular-go-bot/v3/commands"
	"github.com/Gasoid/regular-go-bot/v3/messenger"
)

type Dispatcher struct {
	commands    map[string]commands.Command
	middlewares []func(messenger.Handler) messenger.Handler
}

func (d *Dispatcher) Process(s *messenger.Sender, r messenger.Receiver) error {
	h := processMiddlewares(
		messenger.HandlerFunc(func(sender *messenger.Sender, receiver messenger.Receiver) error {

			return nil
		}), d.middlewares...)

	err := h.Process(s, r)
	if err != nil {
		return err
	}

	return nil
}

func start() {
	var (
		chats      chan messenger.Chat
		dispatcher Dispatcher
	)

	for {

		chat := <-chats

		err := dispatcher.Process(chat.GetSender(), chat.GetReceiver())
		if err != nil {
			slog.Error("handler returns error", "error", err)
		}

	}
}

func processMiddlewares(
	handler messenger.Handler,
	middlewares ...func(messenger.Handler) messenger.Handler) messenger.Handler {

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

func metricsMiddleware(h messenger.Handler) messenger.Handler {
	return messenger.HandlerFunc(
		func(s *messenger.Sender, r messenger.Receiver) error {
			return h.Process(s, r)
		})
}
