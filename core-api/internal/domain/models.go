package domain

import "time"

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleAnalyst  UserRole = "analyst"
	RoleOperator UserRole = "operator"
	RoleViewer   UserRole = "viewer"
)

type Metric string

const (
	MetricTemperature   Metric = "temperature"
	MetricWindSpeed     Metric = "wind_speed"
	MetricHumidity      Metric = "humidity"
	MetricPressure      Metric = "pressure"
	MetricPrecipitation Metric = "precipitation"
)

type WeatherParameter string

const (
	ParameterTemperatureMin WeatherParameter = "temperature_min"
	ParameterTemperatureMax WeatherParameter = "temperature_max"
	ParameterPrecipitation  WeatherParameter = "precipitation_total"
	ParameterWindSpeed      WeatherParameter = "wind_speed"
	ParameterWindGust       WeatherParameter = "wind_gust"
	ParameterHumidity       WeatherParameter = "humidity"
	ParameterPressure       WeatherParameter = "pressure"
)

type StationStatus string

const (
	StationOnline   StationStatus = "online"
	StationDegraded StationStatus = "degraded"
	StationOffline  StationStatus = "offline"
)

type AlertSeverity string
type AlertStatus string
type BackfillStatus string
type TimeBucket string

const (
	AlertCritical AlertSeverity = "critical"
	AlertWarning  AlertSeverity = "warning"
	AlertInfo     AlertSeverity = "info"

	AlertOpen         AlertStatus = "open"
	AlertAcknowledged AlertStatus = "acknowledged"
	AlertResolved     AlertStatus = "resolved"

	BackfillQueued    BackfillStatus = "queued"
	BackfillRunning   BackfillStatus = "running"
	BackfillCompleted BackfillStatus = "completed"
	BackfillFailed    BackfillStatus = "failed"
	BackfillCancelled BackfillStatus = "cancelled"

	Bucket1H TimeBucket = "1h"
	Bucket3H TimeBucket = "3h"
	Bucket6H TimeBucket = "6h"
	Bucket1D TimeBucket = "1d"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Role         UserRole  `json:"role"`
	LastSeen     time.Time `json:"lastSeen"`
	PasswordHash string    `json:"-"`
}

type Region struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Station struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	RegionID        string        `json:"regionId"`
	Lat             float64       `json:"lat"`
	Lon             float64       `json:"lon"`
	Status          StationStatus `json:"status"`
	ActiveSensors   int           `json:"activeSensors"`
	MaxError        float64       `json:"maxError"`
	LastTelemetryAt time.Time     `json:"lastTelemetryAt"`
}

type ForecastReading struct {
	ID         string
	StationID  string
	Metric     Metric
	Value      float64
	ForecastAt time.Time
	TargetAt   time.Time
	Source     string
	RawPayload []byte
}

type ActualWeatherReading struct {
	ID         string
	StationID  string
	ObservedAt time.Time
	Values     map[WeatherParameter]float64
	Source     string
	TraceID    string
	RawPayload []byte
}

type OverviewMetrics struct {
	ActiveStations   int               `json:"activeStations"`
	DegradedStations int               `json:"degradedStations"`
	OfflineStations  int               `json:"offlineStations"`
	KafkaLag         int64             `json:"kafkaLag"`
	RequestRate      float64           `json:"requestRate"`
	P95LatencyMs     float64           `json:"p95LatencyMs"`
	WorstErrorToday  float64           `json:"worstErrorToday"`
	ErrorTrend       []ErrorTrendPoint `json:"errorTrend"`
}

type ErrorTrendPoint struct {
	Date string  `json:"date"`
	MAE  float64 `json:"mae"`
	RMSE float64 `json:"rmse"`
}

type ForecastErrorRow struct {
	ID            string    `json:"id"`
	StationID     string    `json:"stationId"`
	StationName   string    `json:"stationName"`
	RegionName    string    `json:"regionName"`
	Metric        Metric    `json:"metric"`
	ForecastValue float64   `json:"forecastValue"`
	ActualValue   float64   `json:"actualValue"`
	AbsoluteError float64   `json:"absoluteError"`
	ErrorPct      float64   `json:"errorPct"`
	ObservedAt    time.Time `json:"observedAt"`
}

type ParameterErrorRow struct {
	ID              string           `json:"id"`
	Parameter       WeatherParameter `json:"parameter"`
	StationID       string           `json:"stationId"`
	StationName     string           `json:"stationName"`
	RegionName      string           `json:"regionName"`
	ForecastValue   float64          `json:"forecastValue"`
	ActualValue     float64          `json:"actualValue"`
	AbsoluteError   float64          `json:"absoluteError"`
	ErrorPct        float64          `json:"errorPct"`
	ContributionPct float64          `json:"contributionPct"`
	Samples         int              `json:"samples"`
	ObservedAt      time.Time        `json:"observedAt"`
}

type ParameterErrorTrendPoint struct {
	Timestamp     time.Time        `json:"timestamp"`
	Parameter     WeatherParameter `json:"parameter"`
	AbsoluteError float64          `json:"absoluteError"`
	MAE           float64          `json:"mae"`
}

type StationSeriesPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	Forecast      float64   `json:"forecast"`
	Actual        float64   `json:"actual"`
	AbsoluteError float64   `json:"absoluteError"`
}

type HistoricalMetric struct {
	Date            string  `json:"date"`
	RegionName      string  `json:"regionName"`
	StationName     string  `json:"stationName"`
	MAE             float64 `json:"mae"`
	RMSE            float64 `json:"rmse"`
	Samples         int     `json:"samples"`
	BackfillVersion string  `json:"backfillVersion"`
}

type Alert struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Source    string        `json:"source"`
	Severity  AlertSeverity `json:"severity"`
	Status    AlertStatus   `json:"status"`
	StartedAt time.Time     `json:"startedAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

type BackfillJob struct {
	ID                 string         `json:"id"`
	Status             BackfillStatus `json:"status"`
	CalculationVersion string         `json:"calculationVersion"`
	ProgressPct        float64        `json:"progressPct"`
	ProcessedRows      int64          `json:"processedRows"`
	FailedRows         int64          `json:"failedRows"`
	CreatedAt          time.Time      `json:"createdAt"`
	StartedAt          *time.Time     `json:"startedAt"`
	FinishedAt         *time.Time     `json:"finishedAt"`
}

type AnalyticsFilter struct {
	DateFrom  time.Time
	DateTo    time.Time
	RegionID  string
	StationID string
	Metric    Metric
}

type ParameterFilter struct {
	AnalyticsFilter
	Parameter string
	Bucket    TimeBucket
}

type BackfillRequest struct {
	DateFrom           time.Time
	DateTo             time.Time
	RegionID           string
	StationID          string
	Metric             Metric
	CalculationVersion string
}
