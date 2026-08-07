package redisconnect

import (
	"context"
	"fmt"

	"github.com/chishkin-afk/intask/backend/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func Connect(cfg *config.Config) (*redis.Client, error) {
	client := redis.NewClient(getOpts(cfg))
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("can't ping redis: %w", err)
	}

	return client, nil
}

func getOpts(cfg *config.Config) *redis.Options {
	return &redis.Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Cache.Redis.Host,
			cfg.Cache.Redis.Port,
		),
		Username: cfg.Cache.Redis.Auth.Username,
		Password: cfg.Cache.Redis.Auth.Password,
		DB:       cfg.Cache.Redis.Auth.DB,
	}
}
