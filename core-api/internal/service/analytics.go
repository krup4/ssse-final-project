package service

import (
	"context"

	"weather-accuracy/core-api/internal/domain"
)

func (s *Service) Overview(ctx context.Context, filter domain.AnalyticsFilter) (domain.OverviewMetrics, error) {
	return s.deps.Analytics.Overview(ctx, filter)
}

func (s *Service) WorstErrors(ctx context.Context, filter domain.AnalyticsFilter, limit int, sort string) ([]domain.ForecastErrorRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		return nil, domain.ErrValidation
	}
	if sort == "" {
		sort = "absoluteError_desc"
	}
	return s.deps.Analytics.WorstErrors(ctx, filter, limit, sort)
}

func (s *Service) ParameterErrors(ctx context.Context, filter domain.ParameterFilter) ([]domain.ParameterErrorRow, error) {
	return s.deps.Analytics.ParameterErrors(ctx, filter)
}

func (s *Service) ParameterTrend(ctx context.Context, filter domain.ParameterFilter) ([]domain.ParameterErrorTrendPoint, error) {
	if filter.Bucket == "" {
		filter.Bucket = domain.Bucket1H
	}
	return s.deps.Analytics.ParameterTrend(ctx, filter)
}

func (s *Service) StationSeries(ctx context.Context, filter domain.AnalyticsFilter, bucket domain.TimeBucket) ([]domain.StationSeriesPoint, error) {
	if bucket == "" {
		bucket = domain.Bucket1H
	}
	return s.deps.Analytics.StationSeries(ctx, filter, bucket)
}

func (s *Service) History(ctx context.Context, filter domain.AnalyticsFilter) ([]domain.HistoricalMetric, error) {
	return s.deps.Analytics.History(ctx, filter)
}
