package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, redisURL, redisURI, host, port, password string, db int, errorLogger *slog.Logger) (*redis.Client, error) {
	connStr := redisURL
	if connStr == "" {
		connStr = redisURI
	}

	var opts *redis.Options
	var err error

	if connStr != "" {
		opts, err = redis.ParseURL(connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse redis url: %w", err)
		}
	} else {
		opts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: password,
			DB:       db,
		}
	}

	rdb := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		if errorLogger != nil {
			errorLogger.Warn("redis ping failed on startup", "error", err)
		}
		return rdb, fmt.Errorf("redis ping failed: %w", err)
	}

	return rdb, nil
}
