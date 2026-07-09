package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

type Options struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnectRetries  int
	ConnectBackoff  time.Duration
	PingTimeout     time.Duration
}

type AnalyticsRepository struct {
	db *sql.DB
}

func Open(ctx context.Context, options Options) (*sql.DB, error) {
	if options.ConnectRetries <= 0 {
		options.ConnectRetries = 1
	}
	if options.ConnectBackoff <= 0 {
		options.ConnectBackoff = time.Second
	}
	if options.PingTimeout <= 0 {
		options.PingTimeout = 2 * time.Second
	}

	var lastErr error
	backoff := options.ConnectBackoff
	for attempt := 1; attempt <= options.ConnectRetries; attempt++ {
		db, err := openOnce(ctx, options)
		if err == nil {
			return db, nil
		}
		lastErr = err
		if attempt == options.ConnectRetries {
			break
		}
		if !sleep(ctx, backoff) {
			return nil, ctx.Err()
		}
		backoff = minDuration(backoff*2, 30*time.Second)
	}
	return nil, fmt.Errorf("clickhouse connection failed after %d attempts: %w", options.ConnectRetries, lastErr)
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

func (r *AnalyticsRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func openOnce(ctx context.Context, options Options) (*sql.DB, error) {
	db, err := sql.Open("clickhouse", options.DSN)
	if err != nil {
		return nil, err
	}
	if options.MaxOpenConns > 0 {
		db.SetMaxOpenConns(options.MaxOpenConns)
	}
	if options.MaxIdleConns > 0 {
		db.SetMaxIdleConns(options.MaxIdleConns)
	}
	if options.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(options.ConnMaxLifetime)
	}
	pingCtx, cancel := context.WithTimeout(ctx, options.PingTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func sleep(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
