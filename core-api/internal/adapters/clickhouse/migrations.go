package clickhouse

import (
	"context"
	_ "embed"
	"strings"
)

//go:embed schema.sql
var schemaSQL string

func (r *AnalyticsRepository) Migrate(ctx context.Context) error {
	for _, statement := range strings.Split(schemaSQL, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := r.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	if _, err := r.db.ExecContext(ctx, "ALTER TABLE worst_errors ADD COLUMN IF NOT EXISTS parameter LowCardinality(String) AFTER region_name"); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, "ALTER TABLE parameter_errors ADD COLUMN IF NOT EXISTS metric LowCardinality(String) AFTER parameter"); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, "ALTER TABLE parameter_error_trend ADD COLUMN IF NOT EXISTS metric LowCardinality(String) AFTER parameter"); err != nil {
		return err
	}
	return nil
}
