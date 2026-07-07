package httpgin

import (
	"context"
	"log/slog"
	"time"

	"weather-accuracy/core-api/internal/config"
	"weather-accuracy/core-api/internal/platform/metrics"
	"weather-accuracy/core-api/internal/service"
)

type Dependencies struct {
	Config    config.Config
	Logger    *slog.Logger
	Service   *service.Service
	Readiness func(context.Context) error
	ReadyTTL  time.Duration
	Metrics   *metrics.Registry
	Cache     CacheStore
	Limiter   RateLimiter
}

type CacheStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type Handler struct {
	cfg            config.Config
	log            *slog.Logger
	svc            *service.Service
	ready          func(context.Context) error
	readyTTL       time.Duration
	metricRegistry *metrics.Registry
	cache          CacheStore
	limiter        RateLimiter
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		cfg:            deps.Config,
		log:            deps.Logger,
		svc:            deps.Service,
		ready:          deps.Readiness,
		readyTTL:       deps.ReadyTTL,
		metricRegistry: deps.Metrics,
		cache:          deps.Cache,
		limiter:        deps.Limiter,
	}
}
