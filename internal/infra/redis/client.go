package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/HMasataka/gate/internal/config"
)

func NewClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	opt.PoolSize = cfg.PoolSize
	opt.MinIdleConns = cfg.MinIdleConns
	opt.DialTimeout = cfg.DialTimeout
	opt.ReadTimeout = cfg.ReadTimeout
	opt.WriteTimeout = cfg.WriteTimeout

	client := redis.NewClient(opt)

	if err := pingWithRetry(ctx, client, 5); err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	return client, nil
}

func pingWithRetry(ctx context.Context, client *redis.Client, maxRetries int) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		if err = client.Ping(ctx).Err(); err == nil {
			return nil
		}
		if i < maxRetries {
			backoff := min(time.Duration(1<<uint(i))*time.Second, 30*time.Second)
			slog.Warn("redis connection failed, retrying",
				"attempt", i+1, "max_retries", maxRetries, "backoff", backoff, "error", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return fmt.Errorf("after %d retries: %w", maxRetries, err)
}
