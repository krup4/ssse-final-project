package clickhouse

import (
	"context"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"time"

	"gorm.io/gorm"
)

const syncLockKey int64 = 894233517740506112

type syncSourceRow struct {
	ForecastID    int
	MetricID      int
	ID            string
	StationID     string
	StationName   string
	RegionID      string
	RegionName    string
	Parameter     string
	Metric        string
	ForecastValue float64
	MetricValue   float64
	ObservedAt    time.Time
}

type dailyKey struct {
	date        time.Time
	regionID    string
	regionName  string
	stationID   string
	stationName string
	metric      string
}

type dailyAccumulator struct {
	sumAbs     float64
	sumSquared float64
	samples    uint64
	worst      float64
}

func (a *dailyAccumulator) add(absoluteError float64) {
	a.sumAbs += absoluteError
	a.sumSquared += absoluteError * absoluteError
	a.samples++
	if absoluteError > a.worst {
		a.worst = absoluteError
	}
}

func (a dailyAccumulator) mae() float64 {
	if a.samples == 0 {
		return 0
	}
	return a.sumAbs / float64(a.samples)
}

func (a dailyAccumulator) rmse() float64 {
	if a.samples == 0 {
		return 0
	}
	return math.Sqrt(a.sumSquared / float64(a.samples))
}

func (r *AnalyticsRepository) SyncFromPostgres(ctx context.Context, source *gorm.DB) error {
	return source.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		locked, err := tryAdvisoryLock(ctx, tx)
		if err != nil {
			return err
		}
		if !locked {
			return nil
		}

		var rows []syncSourceRow
		if err := tx.Raw(`
		select f.id as forecast_id,
		       m.id as metric_id,
		       concat(f.id::text, '-', m.id::text) as id,
		       f.station_id::text as station_id,
		       s.name as station_name,
		       '' as region_id,
		       '' as region_name,
		       ff.name as parameter,
		       m.name as metric,
		       f.value as forecast_value,
		       a.value as metric_value,
		       a.dt as observed_at
		from forecasts f
		join stations s on s.id = f.station_id and s.is_active = true
		join forecast_fields ff on ff.id = f.field_id
		join metrics m on m.forecast_field_id = ff.id
		join archive a on a.station_id = f.station_id and a.metric_id = m.id and a.dt = f.date
	`).Scan(&rows).Error; err != nil {
			return err
		}

		var activeStations, offlineStations int64
		if err := tx.Table("stations").Where("is_active = true").Count(&activeStations).Error; err != nil {
			return err
		}
		if err := tx.Table("stations").Where("is_active = false").Count(&offlineStations).Error; err != nil {
			return err
		}

		if err := r.truncateAnalyticsTables(ctx); err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return r.insertAnalyticsRows(ctx, rows, uint32(activeStations), uint32(offlineStations))
	})
}

func (r *AnalyticsRepository) truncateAnalyticsTables(ctx context.Context) error {
	for _, table := range []string{"worst_errors", "parameter_errors", "parameter_error_trend", "station_series", "daily_metrics"} {
		if _, err := r.db.ExecContext(ctx, "TRUNCATE TABLE "+table); err != nil {
			return err
		}
	}
	return nil
}

func (r *AnalyticsRepository) insertAnalyticsRows(ctx context.Context, rows []syncSourceRow, activeStations uint32, offlineStations uint32) error {
	totalAbs := 0.0
	for _, row := range rows {
		totalAbs += row.MetricValue
	}

	daily := map[dailyKey]*dailyAccumulator{}
	for _, row := range rows {
		metricValue := row.MetricValue
		displayActual := displayActualValue(row.ForecastValue, row.Metric, metricValue)
		errPct := metricPercent(row.ForecastValue, metricValue)
		contribution := 0.0
		if totalAbs > 0 {
			contribution = metricValue / totalAbs * 100
		}
		id := stableID(row.ID, row.Parameter, row.Metric, row.ObservedAt)
		if _, err := r.db.ExecContext(ctx, `INSERT INTO worst_errors
			(id, station_id, station_name, region_id, region_name, parameter, metric, forecast_value, actual_value, absolute_error, error_pct, observed_at, backfill_version)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, row.StationID, row.StationName, row.RegionID, row.RegionName, row.Parameter, row.Metric, row.ForecastValue, displayActual, metricValue, errPct, row.ObservedAt.UTC(), "calc-v1"); err != nil {
			return fmt.Errorf("insert worst_errors: %w", err)
		}
		if _, err := r.db.ExecContext(ctx, `INSERT INTO parameter_errors
			(id, parameter, metric, station_id, station_name, region_id, region_name, forecast_value, actual_value, absolute_error, error_pct, contribution_pct, samples, observed_at, backfill_version)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, row.Parameter, row.Metric, row.StationID, row.StationName, row.RegionID, row.RegionName, row.ForecastValue, displayActual, metricValue, errPct, contribution, uint64(1), row.ObservedAt.UTC(), "calc-v1"); err != nil {
			return fmt.Errorf("insert parameter_errors: %w", err)
		}
		if _, err := r.db.ExecContext(ctx, `INSERT INTO parameter_error_trend
			(timestamp, parameter, metric, region_id, station_id, absolute_error, mae, samples, backfill_version)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			row.ObservedAt.UTC(), row.Parameter, row.Metric, row.RegionID, row.StationID, metricValue, metricValue, uint64(1), "calc-v1"); err != nil {
			return fmt.Errorf("insert parameter_error_trend: %w", err)
		}
		if _, err := r.db.ExecContext(ctx, `INSERT INTO station_series
			(timestamp, station_id, metric, forecast, actual, absolute_error, backfill_version)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			row.ObservedAt.UTC(), row.StationID, row.Metric, row.ForecastValue, displayActual, metricValue, "calc-v1"); err != nil {
			return fmt.Errorf("insert station_series: %w", err)
		}

		date := time.Date(row.ObservedAt.UTC().Year(), row.ObservedAt.UTC().Month(), row.ObservedAt.UTC().Day(), 0, 0, 0, 0, time.UTC)
		key := dailyKey{date: date, regionID: row.RegionID, regionName: row.RegionName, stationID: row.StationID, stationName: row.StationName, metric: row.Metric}
		acc := daily[key]
		if acc == nil {
			acc = &dailyAccumulator{}
			daily[key] = acc
		}
		acc.add(metricValue)
	}

	keys := make([]dailyKey, 0, len(daily))
	for key := range daily {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].date.Equal(keys[j].date) {
			return keys[i].stationID < keys[j].stationID
		}
		return keys[i].date.Before(keys[j].date)
	})
	for _, key := range keys {
		acc := daily[key]
		if _, err := r.db.ExecContext(ctx, `INSERT INTO daily_metrics
			(date, region_id, region_name, station_id, station_name, metric, mae, rmse, samples, worst_error_today, active_stations, degraded_stations, offline_stations, kafka_lag, request_rate, p95_latency_ms, backfill_version)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			key.date, key.regionID, key.regionName, key.stationID, key.stationName, key.metric, acc.mae(), acc.rmse(), acc.samples, acc.worst, activeStations, uint32(0), offlineStations, int64(0), float64(0), float64(0), "calc-v1"); err != nil {
			return fmt.Errorf("insert daily_metrics: %w", err)
		}
	}
	return nil
}

func tryAdvisoryLock(ctx context.Context, db *gorm.DB) (bool, error) {
	var locked bool
	err := db.WithContext(ctx).Raw("select pg_try_advisory_xact_lock(?)", syncLockKey).Scan(&locked).Error
	return locked, err
}

func metricPercent(forecast, metricValue float64) float64 {
	if math.Abs(forecast) < 0.000001 {
		return 0
	}
	return math.Abs(metricValue / forecast * 100)
}

func displayActualValue(forecast float64, metric string, metricValue float64) float64 {
	if metric == "mse" {
		return forecast - math.Sqrt(math.Abs(metricValue))
	}
	return forecast - metricValue
}

func stableID(parts ...any) string {
	hash := sha1.Sum([]byte(fmt.Sprint(parts...)))
	prefix := binary.BigEndian.Uint64(hash[:8])
	return fmt.Sprintf("%x", prefix)
}
