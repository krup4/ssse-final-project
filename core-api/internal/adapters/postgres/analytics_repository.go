package postgres

import (
	"context"

	"weather-accuracy/core-api/internal/domain"
)

func (r *AnalyticsRepository) Overview(ctx context.Context, filter domain.AnalyticsFilter) (domain.OverviewMetrics, error) {
	var overview domain.OverviewMetrics
	var active, degraded, offline int64
	_ = r.db.WithContext(ctx).Model(&StationModel{}).Where("status = ?", "online").Count(&active).Error
	_ = r.db.WithContext(ctx).Model(&StationModel{}).Where("status = ?", "degraded").Count(&degraded).Error
	_ = r.db.WithContext(ctx).Model(&StationModel{}).Where("status = ?", "offline").Count(&offline).Error
	overview.ActiveStations = int(active)
	overview.DegradedStations = int(degraded)
	overview.OfflineStations = int(offline)
	overview.RequestRate = 0
	overview.P95LatencyMs = 0
	overview.KafkaLag = 0
	rows, err := r.WorstErrors(ctx, filter, 1, "absoluteError_desc")
	if err != nil {
		return overview, err
	}
	if len(rows) > 0 {
		overview.WorstErrorToday = rows[0].AbsoluteError
	}
	err = r.db.WithContext(ctx).Raw(`
		with errors as (`+baseErrorSQL()+`)
		select to_char(date_trunc('day', observed_at), 'YYYY-MM-DD') as date,
		       avg(absolute_error) as mae,
		       sqrt(avg(absolute_error * absolute_error)) as rmse
		from errors
		group by 1
		order by 1`, sqlArgs(filter)...).Scan(&overview.ErrorTrend).Error
	return overview, err
}

func (r *AnalyticsRepository) WorstErrors(ctx context.Context, filter domain.AnalyticsFilter, limit int, sort string) ([]domain.ForecastErrorRow, error) {
	order := "absolute_error desc"
	if sort == "absoluteError_asc" {
		order = "absolute_error asc"
	}
	var rows []domain.ForecastErrorRow
	err := r.db.WithContext(ctx).Raw(`
		with errors as (`+baseErrorSQL()+`)
		select id, station_id, station_name, region_name, metric, forecast_value, actual_value,
		       absolute_error, error_pct, observed_at
		from errors
		order by `+order+`
		limit ?`, append(sqlArgs(filter), limit)...).Scan(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) ParameterErrors(ctx context.Context, filter domain.ParameterFilter) ([]domain.ParameterErrorRow, error) {
	var rows []domain.ParameterErrorRow
	args := parameterArgs(filter)
	err := r.db.WithContext(ctx).Raw(`
		with parameter_errors as (`+baseParameterErrorSQL(filter.Parameter)+`),
		     total as (select nullif(sum(absolute_error), 0) total_error from parameter_errors)
		select id, parameter, station_id, station_name, region_name, forecast_value, actual_value,
		       absolute_error, error_pct,
		       round((absolute_error / total.total_error * 100)::numeric, 2)::float as contribution_pct,
		       1 as samples,
		       observed_at
		from parameter_errors, total
		order by absolute_error desc
		limit 500`, args...).Scan(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) ParameterTrend(ctx context.Context, filter domain.ParameterFilter) ([]domain.ParameterErrorTrendPoint, error) {
	var rows []domain.ParameterErrorTrendPoint
	bucket := bucketExpression(filter.Bucket, "observed_at")
	err := r.db.WithContext(ctx).Raw(`
		with parameter_errors as (`+baseParameterErrorSQL(filter.Parameter)+`)
		select `+bucket+` as timestamp, parameter,
		       max(absolute_error) as absolute_error,
		       avg(absolute_error) as mae
		from parameter_errors
		group by 1, 2
		order by 1, 2`, parameterArgs(filter)...).Scan(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) StationSeries(ctx context.Context, filter domain.AnalyticsFilter, bucket domain.TimeBucket) ([]domain.StationSeriesPoint, error) {
	var rows []domain.StationSeriesPoint
	err := r.db.WithContext(ctx).Raw(`
		with errors as (`+baseErrorSQL()+`)
		select `+bucketExpression(bucket, "observed_at")+` as timestamp,
		       avg(forecast_value) as forecast,
		       avg(actual_value) as actual,
		       avg(absolute_error) as absolute_error
		from errors
		group by 1
		order by 1`, sqlArgs(filter)...).Scan(&rows).Error
	return rows, err
}

func (r *AnalyticsRepository) History(ctx context.Context, filter domain.AnalyticsFilter) ([]domain.HistoricalMetric, error) {
	var rows []domain.HistoricalMetric
	err := r.db.WithContext(ctx).Raw(`
		with errors as (`+baseErrorSQL()+`)
		select to_char(date_trunc('day', observed_at), 'YYYY-MM-DD') as date,
		       region_name, station_name,
		       avg(absolute_error) as mae,
		       sqrt(avg(absolute_error * absolute_error)) as rmse,
		       count(*) as samples,
		       'calc-v1' as backfill_version
		from errors
		group by 1, 2, 3
		order by 1 desc, mae desc`, sqlArgs(filter)...).Scan(&rows).Error
	return rows, err
}
