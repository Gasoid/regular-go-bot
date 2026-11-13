package main

import (
	"slices"
	"time"

	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/Gasoid/regular-go-bot/v3/metrics"
)

func processMiddlewares(
	handler messenger.Handler,
	middlewares ...func(messenger.Handler) messenger.Handler) messenger.Handler {

	for i := range slices.Backward(middlewares) {
		handler = middlewares[i](handler)
	}

	return handler
}

func metricsMiddleware(h messenger.Handler) messenger.Handler {
	return messenger.HandlerFunc(
		func(s *messenger.Sender, r messenger.Receiver) error {
			start := time.Now()
			err := h.Process(s, r)

			switch s.MessageType {
			case messenger.Command:
				metrics.CommandInc(s.Message.Command, err)
				metrics.CommandDuration(s.Message.Command, time.Since(start))
			}

			return err
		})
}
