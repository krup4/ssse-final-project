package monitoring

import (
    "context"
    "fmt"
)

type Checker struct {
    dbHealth   func(context.Context) error
    kafkaHealth func(context.Context) error
    redisHealth func(context.Context) error
}

func NewChecker(db func(context.Context) error, kafka func(context.Context) error, redis func(context.Context) error) *Checker {
    return &Checker{dbHealth: db, kafkaHealth: kafka, redisHealth: redis}
}

func (c *Checker) Check(ctx context.Context) error {
    if err := c.dbHealth(ctx); err != nil {
        return fmt.Errorf("database: %w", err)
    }
    if err := c.kafkaHealth(ctx); err != nil {
        return fmt.Errorf("kafka: %w", err)
    }
    if err := c.redisHealth(ctx); err != nil {
        return fmt.Errorf("redis: %w", err)
    }
    return nil
}
