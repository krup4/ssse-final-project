package config

import (
    "os"
    "time"
)

type Config struct {
    Port        string
    DBUrl       string
    KafkaBrokers string
    RedisUrl    string
    OpenMeteoURL string
    HTTPTimeout time.Duration
}

func Load() Config {
    return Config{
        Port: getEnv("PORT", "8080"),
        DBUrl: getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/weather?sslmode=disable"),
        KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
        RedisUrl: getEnv("REDIS_URL", "redis://localhost:6379"),
        OpenMeteoURL: getEnv("OPEN_METEO_URL", "https://api.open-meteo.com"),
        HTTPTimeout: 10 * time.Second,
    }
}

func getEnv(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
