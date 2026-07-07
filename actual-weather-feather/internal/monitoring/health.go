package monitoring

import (
    "context"
)

type HealthChecker interface {
    Check(ctx context.Context) error
}

type HealthStatus struct {
    Database bool `json:"database"`
    Kafka    bool `json:"kafka"`
    Redis    bool `json:"redis"`
}
