package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     string
	Password string
}

type Client struct {
	client *redis.Client
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Client{client: rdb}, nil
}

// увеличивает счетчик на 1
func (c *Client) Incr(ctx context.Context, key string) error {
	return c.client.Incr(ctx, key).Err()
}

// уменьшает счетчик на 1
func (c *Client) Decr(ctx context.Context, key string) error {
	return c.client.Decr(ctx, key).Err()
}

// возвращает числовое значение ключа
func (c *Client) GetInt(ctx context.Context, key string) (int, error) {
	val, err := c.client.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}

// сохраняет ключ со сроком жизни TTL
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}
