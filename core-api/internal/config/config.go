package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env              string
	HTTPAddr         string
	HTTPRequestTTL   time.Duration
	LogLevel         string
	DatabaseURL      string
	Database         DatabaseConfig
	JWTSecret        string
	TokenTTL         time.Duration
	ShutdownTimeout  time.Duration
	AutoMigrate      bool
	AnalyticsBackend string
	RateLimitRPS     int
	Kafka            KafkaConfig
	Redis            RedisConfig
	ClickHouse       ClickHouseConfig
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectRetries  int
	ConnectBackoff  time.Duration
	ConnectTimeout  time.Duration
	PingTimeout     time.Duration
}

type KafkaConfig struct {
	Enabled            bool
	Brokers            []string
	ActualWeatherTopic string
	BackfillJobsTopic  string
	GroupID            string
	MinBytes           int
	MaxBytes           int
	CommitInterval     time.Duration
}

type RedisConfig struct {
	Enabled      bool
	Addr         string
	Password     string
	DB           int
	KeyPrefix    string
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	CacheTTL     time.Duration
	DedupTTL     time.Duration
	RateLimitTTL time.Duration
}

type ClickHouseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnectRetries  int
	ConnectBackoff  time.Duration
	ConnectTimeout  time.Duration
	PingTimeout     time.Duration
}

func Load() Config {
	return Config{
		Env:            env("APP_ENV", "local"),
		HTTPAddr:       env("HTTP_ADDR", ":8080"),
		HTTPRequestTTL: durationEnv("HTTP_REQUEST_TIMEOUT", 20*time.Second),
		LogLevel:       env("LOG_LEVEL", "info"),
		DatabaseURL:    env("DATABASE_URL", "postgres://weather:weather@localhost:5432/weather_accuracy?sslmode=disable"),
		Database: DatabaseConfig{
			URL:             env("DATABASE_URL", "postgres://weather:weather@localhost:5432/weather_accuracy?sslmode=disable"),
			MaxOpenConns:    intEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    intEnv("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: durationEnv("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: durationEnv("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
			ConnectRetries:  intEnv("DB_CONNECT_RETRIES", 10),
			ConnectBackoff:  durationEnv("DB_CONNECT_BACKOFF", time.Second),
			ConnectTimeout:  durationEnv("DB_CONNECT_TIMEOUT", 2*time.Minute),
			PingTimeout:     durationEnv("DB_PING_TIMEOUT", 2*time.Second),
		},
		JWTSecret:        env("JWT_SECRET", "local-dev-secret-change-me"),
		TokenTTL:         durationEnv("TOKEN_TTL", 8*time.Hour),
		ShutdownTimeout:  durationEnv("SHUTDOWN_TIMEOUT", 10*time.Second),
		AutoMigrate:      boolEnv("AUTO_MIGRATE", false),
		AnalyticsBackend: env("ANALYTICS_BACKEND", "postgres"),
		RateLimitRPS:     intEnv("RATE_LIMIT_RPS", 100),
		Kafka: KafkaConfig{
			Enabled:            boolEnv("KAFKA_ENABLED", true),
			Brokers:            splitEnv("KAFKA_BROKERS", "localhost:9092"),
			ActualWeatherTopic: env("KAFKA_ACTUAL_WEATHER_TOPIC", "actual-weather.raw.v1"),
			BackfillJobsTopic:  env("KAFKA_BACKFILL_JOBS_TOPIC", "backfill.jobs.v1"),
			GroupID:            env("KAFKA_GROUP_ID", "core-api-actual-weather"),
			MinBytes:           intEnv("KAFKA_MIN_BYTES", 1),
			MaxBytes:           intEnv("KAFKA_MAX_BYTES", 10e6),
			CommitInterval:     durationEnv("KAFKA_COMMIT_INTERVAL", time.Second),
		},
		Redis: RedisConfig{
			Enabled:      boolEnv("REDIS_ENABLED", false),
			Addr:         env("REDIS_ADDR", "localhost:6379"),
			Password:     env("REDIS_PASSWORD", ""),
			DB:           intEnv("REDIS_DB", 0),
			KeyPrefix:    env("REDIS_KEY_PREFIX", "weather:core-api"),
			DialTimeout:  durationEnv("REDIS_DIAL_TIMEOUT", 2*time.Second),
			ReadTimeout:  durationEnv("REDIS_READ_TIMEOUT", time.Second),
			WriteTimeout: durationEnv("REDIS_WRITE_TIMEOUT", time.Second),
			CacheTTL:     durationEnv("REDIS_CACHE_TTL", 30*time.Second),
			DedupTTL:     durationEnv("REDIS_DEDUP_TTL", 24*time.Hour),
			RateLimitTTL: durationEnv("REDIS_RATE_LIMIT_TTL", time.Second),
		},
		ClickHouse: ClickHouseConfig{
			DSN:             env("CLICKHOUSE_DSN", "clickhouse://weather:weather@localhost:9000/weather_dwh"),
			MaxOpenConns:    intEnv("CLICKHOUSE_MAX_OPEN_CONNS", 20),
			MaxIdleConns:    intEnv("CLICKHOUSE_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: durationEnv("CLICKHOUSE_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnectRetries:  intEnv("CLICKHOUSE_CONNECT_RETRIES", 10),
			ConnectBackoff:  durationEnv("CLICKHOUSE_CONNECT_BACKOFF", time.Second),
			ConnectTimeout:  durationEnv("CLICKHOUSE_CONNECT_TIMEOUT", 2*time.Minute),
			PingTimeout:     durationEnv("CLICKHOUSE_PING_TIMEOUT", 2*time.Second),
		},
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitEnv(key, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func boolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
