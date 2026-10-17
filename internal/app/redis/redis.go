package redis

import (
	"context"
	"fmt"
	"stellar-measurements-backend/internal/app/config"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "stellar_service." // префикс сервиса для изоляции ключей в Redis

type Client struct {
	cfg    config.RedisConfig
	client *redis.Client
}

func New(ctx context.Context, cfg config.RedisConfig) (*Client, error) {
	client := &Client{}
	client.cfg = cfg

	addr := cfg.Host
	if cfg.Port > 0 {
		addr = addr + ":" + strconv.Itoa(cfg.Port)
	}

	redisClient := redis.NewClient(&redis.Options{
		Password:    cfg.Password,
		Username:    cfg.User,
		Addr:        addr,
		DB:          0,
		DialTimeout: cfg.DialTimeout,
		ReadTimeout: cfg.ReadTimeout,
	})

	client.client = redisClient

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return client, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) RawClient() *redis.Client {
	return c.client
}
