package redisstore

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Options struct {
	Addr         string
	Password     string
	DB           int
	KeyPrefix    string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Client struct {
	client *redis.Client
	prefix string
}

func New(ctx context.Context, options Options) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         options.Addr,
		Password:     options.Password,
		DB:           options.DB,
		DialTimeout:  options.DialTimeout,
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &Client{client: client, prefix: strings.Trim(options.KeyPrefix, ":")}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit <= 0 {
		return true, nil
	}
	if window <= 0 {
		window = time.Second
	}
	redisKey := c.key("rate", key, time.Now().UTC().Truncate(window).Format(time.RFC3339))
	pipe := c.client.TxPipeline()
	count := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, window+time.Second)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return count.Val() <= int64(limit), nil
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := c.client.Get(ctx, c.key("cache", key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return value, err
}

func (c *Client) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return c.client.Set(ctx, c.key("cache", key), value, ttl).Err()
}

func (c *Client) MarkOnce(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if ttl <= 0 {
		return true, nil
	}
	return c.client.SetNX(ctx, c.key("dedup", key), "1", ttl).Result()
}

func (c *Client) key(parts ...string) string {
	clean := make([]string, 0, len(parts)+1)
	if c.prefix != "" {
		clean = append(clean, c.prefix)
	}
	for _, part := range parts {
		part = strings.Trim(part, ":")
		if part != "" {
			clean = append(clean, part)
		}
	}
	return strings.Join(clean, ":")
}
