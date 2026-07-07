package redis

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type Client interface {
    Ping(ctx context.Context) error
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Close() error
}

type redisClient struct {
    client *redis.Client
}

func NewClient(url string) (Client, error) {
    opts, err := redis.ParseURL(url)
    if err != nil {
        return nil, fmt.Errorf("parse redis url: %w", err)
    }

    client := redis.NewClient(opts)
    return &redisClient{client: client}, nil
}

func (r *redisClient) Ping(ctx context.Context) error {
    return r.client.Ping(ctx).Err()
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
    return r.client.Get(ctx, key).Result()
}

func (r *redisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisClient) Close() error {
    return r.client.Close()
}
