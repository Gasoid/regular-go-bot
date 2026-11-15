package kv

import (
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
)

func Init() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		slog.Warn("REDIS_ADDR is empty")
		return
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
}

func GetClient() *redis.Client {
	return rdb
}
