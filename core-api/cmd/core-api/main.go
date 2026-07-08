package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weather-accuracy/core-api/internal/adapters/clickhouse"
	"weather-accuracy/core-api/internal/adapters/httpgin"
	"weather-accuracy/core-api/internal/adapters/kafka"
	"weather-accuracy/core-api/internal/adapters/postgres"
	"weather-accuracy/core-api/internal/adapters/redisstore"
	"weather-accuracy/core-api/internal/config"
	"weather-accuracy/core-api/internal/domain"
	"weather-accuracy/core-api/internal/platform/logger"
	"weather-accuracy/core-api/internal/platform/metrics"
	"weather-accuracy/core-api/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)
	metricRegistry := metrics.NewRegistry()

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	defer cancelStartup()

	db, err := postgres.Open(startupCtx, postgres.Options{
		URL:             cfg.Database.URL,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		ConnMaxIdleTime: cfg.Database.ConnMaxIdleTime,
		ConnectRetries:  cfg.Database.ConnectRetries,
		ConnectBackoff:  cfg.Database.ConnectBackoff,
		PingTimeout:     cfg.Database.PingTimeout,
	})
	if err != nil {
		log.Error("postgres connection failed", slog.Any("error", err))
		os.Exit(1)
	}

	repos := postgres.NewRepositories(db)
	if cfg.AutoMigrate {
		if err := repos.Migrate(context.Background()); err != nil {
			log.Error("postgres migration failed", slog.Any("error", err))
			os.Exit(1)
		}
		if err := repos.SeedDemoData(context.Background()); err != nil {
			log.Error("demo seed failed", slog.Any("error", err))
			os.Exit(1)
		}
	}

	analyticsRepo := domain.AnalyticsRepository(repos.Analytics)
	readiness := repos.Ping
	var clickHouseAnalytics *clickhouse.AnalyticsRepository
	if cfg.AnalyticsBackend == "clickhouse" {
		clickHouseCtx, cancelClickHouse := context.WithTimeout(context.Background(), cfg.ClickHouse.ConnectTimeout)
		chDB, err := clickhouse.Open(clickHouseCtx, clickhouse.Options{
			DSN:             cfg.ClickHouse.DSN,
			MaxOpenConns:    cfg.ClickHouse.MaxOpenConns,
			MaxIdleConns:    cfg.ClickHouse.MaxIdleConns,
			ConnMaxLifetime: cfg.ClickHouse.ConnMaxLifetime,
			ConnectRetries:  cfg.ClickHouse.ConnectRetries,
			ConnectBackoff:  cfg.ClickHouse.ConnectBackoff,
			PingTimeout:     cfg.ClickHouse.PingTimeout,
		})
		cancelClickHouse()
		if err != nil {
			log.Error("clickhouse connection failed", slog.Any("error", err))
			os.Exit(1)
		}
		defer func() {
			if err := chDB.Close(); err != nil {
				log.Warn("clickhouse close failed", slog.Any("error", err))
			}
		}()
		clickHouseAnalytics = clickhouse.NewAnalyticsRepository(chDB)
		analyticsRepo = clickHouseAnalytics
		readiness = func(ctx context.Context) error {
			if err := repos.Ping(ctx); err != nil {
				return err
			}
			return clickHouseAnalytics.Ping(ctx)
		}
	}

	var redisClient *redisstore.Client
	if cfg.Redis.Enabled {
		client, err := redisstore.New(startupCtx, redisstore.Options{
			Addr:         cfg.Redis.Addr,
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.DB,
			KeyPrefix:    cfg.Redis.KeyPrefix,
			DialTimeout:  cfg.Redis.DialTimeout,
			ReadTimeout:  cfg.Redis.ReadTimeout,
			WriteTimeout: cfg.Redis.WriteTimeout,
		})
		if err != nil {
			log.Warn("redis connection failed, continuing without redis-backed cache/rate-limit/dedup", slog.Any("error", err))
		} else {
			redisClient = client
			defer func() {
				if err := redisClient.Close(); err != nil {
					log.Warn("redis close failed", slog.Any("error", err))
				}
			}()
		}
	}

	var backfillProducer *kafka.BackfillProducer
	if cfg.Kafka.Enabled {
		backfillProducer = kafka.NewBackfillProducer(kafka.ProducerConfig{
			Brokers: cfg.Kafka.Brokers,
			Topic:   cfg.Kafka.BackfillJobsTopic,
		}, metricRegistry)
		defer func() {
			if err := backfillProducer.Close(); err != nil {
				log.Warn("backfill kafka producer close failed", slog.Any("error", err))
			}
		}()
	}

	services := service.New(service.Dependencies{
		Users:             repos.Users,
		Regions:           repos.Regions,
		ForecastFields:    repos.ForecastFields,
		Stations:          repos.Stations,
		Analytics:         analyticsRepo,
		Alerts:            repos.Alerts,
		Backfills:         repos.Backfills,
		BackfillPublisher: backfillProducer,
		JWTSecret:         cfg.JWTSecret,
		TokenTTL:          cfg.TokenTTL,
	})

	router := httpgin.NewRouter(httpgin.Dependencies{
		Config:    cfg,
		Logger:    log,
		Service:   services,
		Readiness: readiness,
		ReadyTTL:  cfg.Database.PingTimeout,
		Metrics:   metricRegistry,
		Cache:     redisClient,
		Limiter:   redisClient,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.Kafka.Enabled {
		consumer := kafka.NewActualWeatherConsumer(kafka.ConsumerConfig{
			Brokers:        cfg.Kafka.Brokers,
			Topic:          cfg.Kafka.ActualWeatherTopic,
			GroupID:        cfg.Kafka.GroupID,
			MinBytes:       cfg.Kafka.MinBytes,
			MaxBytes:       cfg.Kafka.MaxBytes,
			CommitInterval: cfg.Kafka.CommitInterval,
			DedupTTL:       cfg.Redis.DedupTTL,
		}, kafka.NewJSONActualWeatherDecoder(), repos.ActualWeather, log, metricRegistry, redisClient)
		go consumer.Run(ctx)
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("core-api started", slog.String("addr", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", slog.Any("error", err))
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("http shutdown failed", slog.Any("error", err))
	}
}
