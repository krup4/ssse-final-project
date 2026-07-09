package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"weather-accuracy/core-api/internal/domain"
)

func (r *AnalyticsRepository) Overview(ctx context.Context, filter domain.AnalyticsFilter) (domain.OverviewMetrics, error) {
	var overview domain.OverviewMetrics
	err := r.db.QueryRowContext(ctx, `
		SELECT
			toInt32(max(active_stations)),
			toInt32(max(degraded_stations)),
			toInt32(max(offline_stations)),
			toInt64(max(kafka_lag)),
			toFloat64(max(request_rate)),
			toFloat64(max(p95_latency_ms)),
			toFloat64(max(worst_error_today))
		FROM daily_metrics
		WHERE date >= toDate(?) AND date <= toDate(?)
		  AND (? = '' OR ? = 'all' OR region_id = ?)
		  AND (? = '' OR ? = 'all' OR station_id = ?)
		  AND (? = '' OR ? = 'all' OR metric = ?)
	`, filter.DateFrom, filter.DateTo, filter.RegionID, filter.RegionID, filter.RegionID, filter.StationID, filter.StationID, filter.StationID, string(filter.Metric), string(filter.Metric), string(filter.Metric)).
		Scan(&overview.ActiveStations, &overview.DegradedStations, &overview.OfflineStations, &overview.KafkaLag, &overview.RequestRate, &overview.P95LatencyMs, &overview.WorstErrorToday)
	if err != nil {
		return overview, err
	}
	trend, err := r.errorTrend(ctx, filter)
	if err != nil {
		return overview, err
	}
	overview.ErrorTrend = trend
	return overview, nil
}

func (r *AnalyticsRepository) WorstErrors(ctx context.Context, filter domain.AnalyticsFilter, limit int, sort string) ([]domain.ForecastErrorRow, error) {
	order := "absolute_error DESC"
	if sort == "absoluteError_asc" {
		order = "absolute_error ASC"
	}
	query := `
		SELECT id, station_id, station_name, region_name, parameter, metric, forecast_value, actual_value,
		       absolute_error, error_pct, observed_at
		FROM worst_errors
		WHERE observed_at >= ? AND observed_at <= ?
		  AND (? = '' OR ? = 'all' OR region_id = ?)
		  AND (? = '' OR ? = 'all' OR station_id = ?)
		  AND (? = '' OR ? = 'all' OR parameter = ?)
		  AND (? = '' OR ? = 'all' OR metric = ?)
		ORDER BY ` + order + `
		LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, filter.DateFrom, filter.DateTo, filter.RegionID, filter.RegionID, filter.RegionID, filter.StationID, filter.StationID, filter.StationID, filter.Field, filter.Field, filter.Field, string(filter.Metric), string(filter.Metric), string(filter.Metric), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ForecastErrorRow, 0)
	for rows.Next() {
		var row domain.ForecastErrorRow
		var metric string
		if err := rows.Scan(&row.ID, &row.StationID, &row.StationName, &row.RegionName, &row.Parameter, &metric, &row.ForecastValue, &row.ActualValue, &row.AbsoluteError, &row.ErrorPct, &row.ObservedAt); err != nil {
			return nil, err
		}
		row.Metric = domain.Metric(metric)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) ParameterErrors(ctx context.Context, filter domain.ParameterFilter) ([]domain.ParameterErrorRow, error) {
	query := `
		SELECT id, parameter, station_id, station_name, region_name, forecast_value, actual_value,
		       absolute_error, error_pct, contribution_pct, samples, observed_at
		FROM parameter_errors
		WHERE observed_at >= ? AND observed_at <= ?
		  AND (? = '' OR ? = 'all' OR region_id = ?)
		  AND (? = '' OR ? = 'all' OR station_id = ?)
		  AND (? = '' OR ? = 'all' OR parameter = ?)
		  AND (? = '' OR ? = 'all' OR parameter = ?)
		ORDER BY absolute_error DESC
		LIMIT 500`
	rows, err := r.db.QueryContext(ctx, query, filter.DateFrom, filter.DateTo, filter.RegionID, filter.RegionID, filter.RegionID, filter.StationID, filter.StationID, filter.StationID, string(filter.Metric), string(filter.Metric), string(filter.Metric), filter.Parameter, filter.Parameter, filter.Parameter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ParameterErrorRow, 0)
	for rows.Next() {
		var row domain.ParameterErrorRow
		var parameter string
		if err := rows.Scan(&row.ID, &parameter, &row.StationID, &row.StationName, &row.RegionName, &row.ForecastValue, &row.ActualValue, &row.AbsoluteError, &row.ErrorPct, &row.ContributionPct, &row.Samples, &row.ObservedAt); err != nil {
			return nil, err
		}
		row.Parameter = domain.WeatherParameter(parameter)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) ParameterTrend(ctx context.Context, filter domain.ParameterFilter) ([]domain.ParameterErrorTrendPoint, error) {
	bucket := clickHouseBucket(filter.Bucket, "timestamp")
	query := fmt.Sprintf(`
		SELECT %s AS bucket_ts, parameter, max(absolute_error), avg(mae)
		FROM parameter_error_trend
		WHERE timestamp >= ? AND timestamp <= ?
		  AND (? = '' OR ? = 'all' OR region_id = ?)
		  AND (? = '' OR ? = 'all' OR station_id = ?)
		  AND (? = '' OR ? = 'all' OR parameter = ?)
		  AND (? = '' OR ? = 'all' OR parameter = ?)
		GROUP BY bucket_ts, parameter
		ORDER BY bucket_ts, parameter`, bucket)
	rows, err := r.db.QueryContext(ctx, query, filter.DateFrom, filter.DateTo, filter.RegionID, filter.RegionID, filter.RegionID, filter.StationID, filter.StationID, filter.StationID, string(filter.Metric), string(filter.Metric), string(filter.Metric), filter.Parameter, filter.Parameter, filter.Parameter)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ParameterErrorTrendPoint, 0)
	for rows.Next() {
		var row domain.ParameterErrorTrendPoint
		var parameter string
		if err := rows.Scan(&row.Timestamp, &parameter, &row.AbsoluteError, &row.MAE); err != nil {
			return nil, err
		}
		row.Parameter = domain.WeatherParameter(parameter)
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) StationSeries(ctx context.Context, filter domain.AnalyticsFilter, bucket domain.TimeBucket) ([]domain.StationSeriesPoint, error) {
	bucketExpr := clickHouseBucket(bucket, "timestamp")
	query := fmt.Sprintf(`
		SELECT %s AS bucket_ts, avg(forecast), avg(actual), avg(absolute_error)
		FROM station_series
		WHERE timestamp >= ? AND timestamp <= ?
		  AND (? = '' OR ? = 'all' OR station_id = ?)
		  AND (? = '' OR ? = 'all' OR metric = ?)
		GROUP BY bucket_ts
		ORDER BY bucket_ts`, bucketExpr)
	rows, err := r.db.QueryContext(ctx, query, filter.DateFrom, filter.DateTo, filter.StationID, filter.StationID, filter.StationID, string(filter.Metric), string(filter.Metric), string(filter.Metric))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.StationSeriesPoint, 0)
	for rows.Next() {
		var row domain.StationSeriesPoint
		if err := rows.Scan(&row.Timestamp, &row.Forecast, &row.Actual, &row.AbsoluteError); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) History(ctx context.Context, filter domain.AnalyticsFilter) ([]domain.HistoricalMetric, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT toString(date), region_name, station_name, mae, rmse, samples, backfill_version
		FROM daily_metrics
		WHERE date >= toDate(?) AND date <= toDate(?)
		  AND (? = '' OR ? = 'all' OR region_id = ?)
		  AND (? = '' OR ? = 'all' OR station_id = ?)
		  AND (? = '' OR ? = 'all' OR metric = ?)
		ORDER BY date DESC, mae DESC
	`, filter.DateFrom, filter.DateTo, filter.RegionID, filter.RegionID, filter.RegionID, filter.StationID, filter.StationID, filter.StationID, string(filter.Metric), string(filter.Metric), string(filter.Metric))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.HistoricalMetric, 0)
	for rows.Next() {
		var row domain.HistoricalMetric
		if err := rows.Scan(&row.Date, &row.RegionName, &row.StationName, &row.MAE, &row.RMSE, &row.Samples, &row.BackfillVersion); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *AnalyticsRepository) errorTrend(ctx context.Context, filter domain.AnalyticsFilter) ([]domain.ErrorTrendPoint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT toString(date), avg(mae), sqrt(avg(rmse * rmse))
		FROM daily_metrics
		WHERE date >= toDate(?) AND date <= toDate(?)
		  AND (? = '' OR ? = 'all' OR region_id = ?)
		  AND (? = '' OR ? = 'all' OR station_id = ?)
		  AND (? = '' OR ? = 'all' OR metric = ?)
		GROUP BY date
		ORDER BY date
	`, filter.DateFrom, filter.DateTo, filter.RegionID, filter.RegionID, filter.RegionID, filter.StationID, filter.StationID, filter.StationID, string(filter.Metric), string(filter.Metric), string(filter.Metric))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ErrorTrendPoint, 0)
	for rows.Next() {
		var row domain.ErrorTrendPoint
		if err := rows.Scan(&row.Date, &row.MAE, &row.RMSE); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func clickHouseBucket(bucket domain.TimeBucket, column string) string {
	column = strings.TrimSpace(column)
	switch bucket {
	case domain.Bucket3H:
		return "toStartOfInterval(" + column + ", INTERVAL 3 HOUR)"
	case domain.Bucket6H:
		return "toStartOfInterval(" + column + ", INTERVAL 6 HOUR)"
	case domain.Bucket1D:
		return "toStartOfDay(" + column + ")"
	default:
		return "toStartOfHour(" + column + ")"
	}
}
