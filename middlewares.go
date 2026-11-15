package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Gasoid/regular-go-bot/v3/messenger"
	"github.com/Gasoid/regular-go-bot/v3/metrics"
	"github.com/redis/go-redis/v9"

	"github.com/go-redis/redis_rate/v10"
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

func normalizeMiddleware(h messenger.Handler) messenger.Handler {
	return messenger.HandlerFunc(
		func(s *messenger.Sender, r messenger.Receiver) error {
			s.Message.Text = strings.TrimSpace(s.Message.Text)
			s.Message.ReplyToMessage = strings.TrimSpace(s.Message.ReplyToMessage)
			s.Message.Arg = strings.TrimSpace(s.Message.Arg)
			return h.Process(s, r)
		})
}

func rateLimitMiddleware(h messenger.Handler) messenger.Handler {
	const perMinuteLimit int = 5

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		slog.Info("rate limiter has not been set up, because REDIS_ADDR is empty")
		return messenger.HandlerFunc(func(s *messenger.Sender, r messenger.Receiver) error {
			return h.Process(s, r)
		})
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	limiter := redis_rate.NewLimiter(rdb)

	return messenger.HandlerFunc(func(s *messenger.Sender, r messenger.Receiver) error {
		if !isRateLimited(s.MessageType) {
			return h.Process(s, r)
		}

		res, err := limiter.Allow(context.Background(), s.ServiceName+":user:"+s.UserID, redis_rate.PerMinute(perMinuteLimit))
		if err != nil {
			slog.Error("rate limiter unavailable", "err", err)
			return h.Process(s, r)
		}

		if res.Allowed == 0 {
			if s.MessageType == messenger.Command {
				r.SendMessage(fmt.Sprintf("Слишком часто. Попробуйте через %s.", res.RetryAfter.Round(time.Second)))
			}
			return nil
		}

		return h.Process(s, r)
	})
}

func isRateLimited(e messenger.EventType) bool {
	return e == messenger.Command ||
		e == messenger.Mentioned ||
		e == messenger.ReplyMessage
}
