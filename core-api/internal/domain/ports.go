package domain

import (
	"context"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
	List(ctx context.Context) ([]User, error)
	UpdateRole(ctx context.Context, id string, role UserRole) (User, error)
	TouchLastSeen(ctx context.Context, id string) error
}

type RegionRepository interface {
	List(ctx context.Context) ([]Region, error)
}

type StationRepository interface {
	List(ctx context.Context, regionID string, status StationStatus) ([]Station, error)
}

type AnalyticsRepository interface {
	Overview(ctx context.Context, filter AnalyticsFilter) (OverviewMetrics, error)
	WorstErrors(ctx context.Context, filter AnalyticsFilter, limit int, sort string) ([]ForecastErrorRow, error)
	ParameterErrors(ctx context.Context, filter ParameterFilter) ([]ParameterErrorRow, error)
	ParameterTrend(ctx context.Context, filter ParameterFilter) ([]ParameterErrorTrendPoint, error)
	StationSeries(ctx context.Context, filter AnalyticsFilter, bucket TimeBucket) ([]StationSeriesPoint, error)
	History(ctx context.Context, filter AnalyticsFilter) ([]HistoricalMetric, error)
}

type AlertRepository interface {
	List(ctx context.Context, status AlertStatus, severity AlertSeverity, source string) ([]Alert, error)
	UpdateStatus(ctx context.Context, id string, status AlertStatus) (Alert, error)
}

type BackfillRepository interface {
	Create(ctx context.Context, request BackfillRequest) (BackfillJob, error)
	Get(ctx context.Context, id string) (BackfillJob, error)
	HasActiveConflict(ctx context.Context, request BackfillRequest) (bool, error)
}

type BackfillPublisher interface {
	PublishBackfillRequested(ctx context.Context, job BackfillJob, request BackfillRequest) error
}

type ActualWeatherRepository interface {
	SaveReading(ctx context.Context, reading ActualWeatherReading) error
}
