package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "strings"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/segmentio/kafka-go"

    "actual-weather-feather/internal/api"
    "actual-weather-feather/internal/client/openmeteo"
    "actual-weather-feather/internal/config"
    "actual-weather-feather/internal/migrations"
    "actual-weather-feather/internal/monitoring"
    "actual-weather-feather/internal/redis"
    "actual-weather-feather/internal/repository"
    ikafka "actual-weather-feather/internal/kafka"
    "actual-weather-feather/internal/scheduler"
    wservice "actual-weather-feather/internal/service/weather"
)

func main() {
    migrateFlag := flag.Bool("migrate", false, "apply database migrations and exit")
    flag.Parse()

    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    cfg := config.Load()

    log.Printf("starting actual-weather service on port %s", cfg.Port)

    // Database
    dbpool, err := pgxpool.New(context.Background(), cfg.DBUrl)
    if err != nil {
        log.Fatalf("failed to connect db: %v", err)
    }
    defer dbpool.Close()

    if *migrateFlag {
        if err := migrations.Apply(ctx, dbpool); err != nil {
            log.Fatalf("failed to apply migrations: %v", err)
        }
        log.Println("migrations applied successfully")
        return
    }

    // Repositories and clients
    repo := repository.NewPgxRepository(dbpool)
    omClient := openmeteo.NewClient(cfg.OpenMeteoURL, cfg.HTTPTimeout)

    // Redis client
    redisClient, err := redis.NewClient(cfg.RedisUrl)
    if err != nil {
        log.Fatalf("failed to connect redis: %v", err)
    }
    defer redisClient.Close()

    // Kafka publisher
    brokers := strings.Split(cfg.KafkaBrokers, ",")
    publisher := ikafka.NewPublisher(brokers, "weather.actual")
    defer publisher.Close()

    // Monitoring
    metrics := monitoring.New()

    // Service
    weatherSvc := wservice.NewService(omClient, repo, repo, publisher, redisClient, metrics)

    // Scheduler
    sched := scheduler.NewScheduler(weatherSvc, repo, metrics)
    go sched.Start(ctx)

    // Health checks
    checker := monitoring.NewChecker(
        func(ctx context.Context) error { return dbpool.Ping(ctx) },
        func(ctx context.Context) error { return kafkaHealth(ctx, brokers) },
        redisClient.Ping,
    )

    // Router
    r := chi.NewRouter()
    api.RegisterRoutes(r, repo, api.NewHealthHandler(checker))

    // Metrics endpoint
    r.Handle("/metrics", promhttp.Handler())

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%s", cfg.Port),
        Handler: r,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("http server error: %v", err)
        }
    }()

    <-ctx.Done()

    log.Println("shutting down")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = srv.Shutdown(ctx)
}

func kafkaHealth(ctx context.Context, brokers []string) error {
    if len(brokers) == 0 {
        return fmt.Errorf("no kafka brokers configured")
    }

    conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
    if err != nil {
        return err
    }
    return conn.Close()
}
