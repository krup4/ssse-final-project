package kafka

import (
    "context"
    "encoding/json"
    "time"

    "github.com/segmentio/kafka-go"
)

type Publisher interface {
    Publish(ctx context.Context, key string, value any) error
    Close() error
}

type kafkaPublisher struct {
    writer *kafka.Writer
}

func NewPublisher(brokers []string, topic string) Publisher {
    return &kafkaPublisher{
        writer: &kafka.Writer{
            Addr:     kafka.TCP(brokers...),
            Topic:    topic,
            Balancer: &kafka.LeastBytes{},
            Async:    false,
        },
    }
}

func (p *kafkaPublisher) Publish(ctx context.Context, key string, value any) error {
    b, err := json.Marshal(value)
    if err != nil {
        return err
    }
    msg := kafka.Message{
        Key:   []byte(key),
        Value: b,
        Time:  time.Now(),
    }
    return p.writer.WriteMessages(ctx, msg)
}

func (p *kafkaPublisher) Close() error {
    return p.writer.Close()
}
