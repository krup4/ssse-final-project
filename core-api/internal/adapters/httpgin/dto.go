package httpgin

import (
	"time"

	"weather-accuracy/core-api/internal/domain"
)

type userResponse struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Email    string          `json:"email"`
	Role     domain.UserRole `json:"role"`
	LastSeen time.Time       `json:"lastSeen"`
}

type loginResponse struct {
	Token string       `json:"token"`
	User  userResponse `json:"user"`
}

type regionResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type stationResponse struct {
	ID              string               `json:"id"`
	Name            string               `json:"name"`
	RegionID        string               `json:"regionId"`
	Lat             float64              `json:"lat"`
	Lon             float64              `json:"lon"`
	Status          domain.StationStatus `json:"status"`
	ActiveSensors   int                  `json:"activeSensors"`
	MaxError        float64              `json:"maxError"`
	LastTelemetryAt time.Time            `json:"lastTelemetryAt"`
}

type overviewMetricsResponse struct {
	ActiveStations   int                       `json:"activeStations"`
	DegradedStations int                       `json:"degradedStations"`
	OfflineStations  int                       `json:"offlineStations"`
	KafkaLag         int64                     `json:"kafkaLag"`
	RequestRate      float64                   `json:"requestRate"`
	P95LatencyMs     float64                   `json:"p95LatencyMs"`
	WorstErrorToday  float64                   `json:"worstErrorToday"`
	ErrorTrend       []errorTrendPointResponse `json:"errorTrend"`
}

type errorTrendPointResponse struct {
	Date string  `json:"date"`
	MAE  float64 `json:"mae"`
	RMSE float64 `json:"rmse"`
}

type forecastErrorRowResponse struct {
	ID            string        `json:"id"`
	StationID     string        `json:"stationId"`
	StationName   string        `json:"stationName"`
	RegionName    string        `json:"regionName"`
	Metric        domain.Metric `json:"metric"`
	ForecastValue float64       `json:"forecastValue"`
	ActualValue   float64       `json:"actualValue"`
	AbsoluteError float64       `json:"absoluteError"`
	ErrorPct      float64       `json:"errorPct"`
	ObservedAt    time.Time     `json:"observedAt"`
}

type parameterErrorRowResponse struct {
	ID              string                  `json:"id"`
	Parameter       domain.WeatherParameter `json:"parameter"`
	StationID       string                  `json:"stationId"`
	StationName     string                  `json:"stationName"`
	RegionName      string                  `json:"regionName"`
	ForecastValue   float64                 `json:"forecastValue"`
	ActualValue     float64                 `json:"actualValue"`
	AbsoluteError   float64                 `json:"absoluteError"`
	ErrorPct        float64                 `json:"errorPct"`
	ContributionPct float64                 `json:"contributionPct"`
	Samples         int                     `json:"samples"`
	ObservedAt      time.Time               `json:"observedAt"`
}

type parameterErrorTrendPointResponse struct {
	Timestamp     time.Time               `json:"timestamp"`
	Parameter     domain.WeatherParameter `json:"parameter"`
	AbsoluteError float64                 `json:"absoluteError"`
	MAE           float64                 `json:"mae"`
}

type stationSeriesPointResponse struct {
	Timestamp     time.Time `json:"timestamp"`
	Forecast      float64   `json:"forecast"`
	Actual        float64   `json:"actual"`
	AbsoluteError float64   `json:"absoluteError"`
}

type historicalMetricResponse struct {
	Date            string  `json:"date"`
	RegionName      string  `json:"regionName"`
	StationName     string  `json:"stationName"`
	MAE             float64 `json:"mae"`
	RMSE            float64 `json:"rmse"`
	Samples         int     `json:"samples"`
	BackfillVersion string  `json:"backfillVersion"`
}

type alertResponse struct {
	ID        string               `json:"id"`
	Title     string               `json:"title"`
	Source    string               `json:"source"`
	Severity  domain.AlertSeverity `json:"severity"`
	Status    domain.AlertStatus   `json:"status"`
	StartedAt time.Time            `json:"startedAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

type backfillJobResponse struct {
	ID                 string                `json:"id"`
	Status             domain.BackfillStatus `json:"status"`
	CalculationVersion string                `json:"calculationVersion"`
	ProgressPct        float64               `json:"progressPct"`
	ProcessedRows      int64                 `json:"processedRows"`
	FailedRows         int64                 `json:"failedRows"`
	CreatedAt          time.Time             `json:"createdAt"`
	StartedAt          *time.Time            `json:"startedAt"`
	FinishedAt         *time.Time            `json:"finishedAt"`
}

func toUserResponse(user domain.User) userResponse {
	return userResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		LastSeen: user.LastSeen,
	}
}

func toUserResponses(users []domain.User) []userResponse {
	out := make([]userResponse, 0, len(users))
	for _, user := range users {
		out = append(out, toUserResponse(user))
	}
	return out
}

func toRegionResponses(regions []domain.Region) []regionResponse {
	out := make([]regionResponse, 0, len(regions))
	for _, region := range regions {
		out = append(out, regionResponse{ID: region.ID, Name: region.Name})
	}
	return out
}

func toStationResponses(stations []domain.Station) []stationResponse {
	out := make([]stationResponse, 0, len(stations))
	for _, station := range stations {
		out = append(out, stationResponse{
			ID:              station.ID,
			Name:            station.Name,
			RegionID:        station.RegionID,
			Lat:             station.Lat,
			Lon:             station.Lon,
			Status:          station.Status,
			ActiveSensors:   station.ActiveSensors,
			MaxError:        station.MaxError,
			LastTelemetryAt: station.LastTelemetryAt,
		})
	}
	return out
}

func toOverviewMetricsResponse(metrics domain.OverviewMetrics) overviewMetricsResponse {
	return overviewMetricsResponse{
		ActiveStations:   metrics.ActiveStations,
		DegradedStations: metrics.DegradedStations,
		OfflineStations:  metrics.OfflineStations,
		KafkaLag:         metrics.KafkaLag,
		RequestRate:      metrics.RequestRate,
		P95LatencyMs:     metrics.P95LatencyMs,
		WorstErrorToday:  metrics.WorstErrorToday,
		ErrorTrend:       toErrorTrendPointResponses(metrics.ErrorTrend),
	}
}

func toErrorTrendPointResponses(points []domain.ErrorTrendPoint) []errorTrendPointResponse {
	out := make([]errorTrendPointResponse, 0, len(points))
	for _, point := range points {
		out = append(out, errorTrendPointResponse{Date: point.Date, MAE: point.MAE, RMSE: point.RMSE})
	}
	return out
}

func toForecastErrorRowResponses(rows []domain.ForecastErrorRow) []forecastErrorRowResponse {
	out := make([]forecastErrorRowResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, forecastErrorRowResponse{
			ID:            row.ID,
			StationID:     row.StationID,
			StationName:   row.StationName,
			RegionName:    row.RegionName,
			Metric:        row.Metric,
			ForecastValue: row.ForecastValue,
			ActualValue:   row.ActualValue,
			AbsoluteError: row.AbsoluteError,
			ErrorPct:      row.ErrorPct,
			ObservedAt:    row.ObservedAt,
		})
	}
	return out
}

func toParameterErrorRowResponses(rows []domain.ParameterErrorRow) []parameterErrorRowResponse {
	out := make([]parameterErrorRowResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, parameterErrorRowResponse{
			ID:              row.ID,
			Parameter:       row.Parameter,
			StationID:       row.StationID,
			StationName:     row.StationName,
			RegionName:      row.RegionName,
			ForecastValue:   row.ForecastValue,
			ActualValue:     row.ActualValue,
			AbsoluteError:   row.AbsoluteError,
			ErrorPct:        row.ErrorPct,
			ContributionPct: row.ContributionPct,
			Samples:         row.Samples,
			ObservedAt:      row.ObservedAt,
		})
	}
	return out
}

func toParameterErrorTrendPointResponses(points []domain.ParameterErrorTrendPoint) []parameterErrorTrendPointResponse {
	out := make([]parameterErrorTrendPointResponse, 0, len(points))
	for _, point := range points {
		out = append(out, parameterErrorTrendPointResponse{
			Timestamp:     point.Timestamp,
			Parameter:     point.Parameter,
			AbsoluteError: point.AbsoluteError,
			MAE:           point.MAE,
		})
	}
	return out
}

func toStationSeriesPointResponses(points []domain.StationSeriesPoint) []stationSeriesPointResponse {
	out := make([]stationSeriesPointResponse, 0, len(points))
	for _, point := range points {
		out = append(out, stationSeriesPointResponse{
			Timestamp:     point.Timestamp,
			Forecast:      point.Forecast,
			Actual:        point.Actual,
			AbsoluteError: point.AbsoluteError,
		})
	}
	return out
}

func toHistoricalMetricResponses(rows []domain.HistoricalMetric) []historicalMetricResponse {
	out := make([]historicalMetricResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, historicalMetricResponse{
			Date:            row.Date,
			RegionName:      row.RegionName,
			StationName:     row.StationName,
			MAE:             row.MAE,
			RMSE:            row.RMSE,
			Samples:         row.Samples,
			BackfillVersion: row.BackfillVersion,
		})
	}
	return out
}

func toAlertResponse(alert domain.Alert) alertResponse {
	return alertResponse{
		ID:        alert.ID,
		Title:     alert.Title,
		Source:    alert.Source,
		Severity:  alert.Severity,
		Status:    alert.Status,
		StartedAt: alert.StartedAt,
		UpdatedAt: alert.UpdatedAt,
	}
}

func toAlertResponses(alerts []domain.Alert) []alertResponse {
	out := make([]alertResponse, 0, len(alerts))
	for _, alert := range alerts {
		out = append(out, toAlertResponse(alert))
	}
	return out
}

func toBackfillJobResponse(job domain.BackfillJob) backfillJobResponse {
	return backfillJobResponse{
		ID:                 job.ID,
		Status:             job.Status,
		CalculationVersion: job.CalculationVersion,
		ProgressPct:        job.ProgressPct,
		ProcessedRows:      job.ProcessedRows,
		FailedRows:         job.FailedRows,
		CreatedAt:          job.CreatedAt,
		StartedAt:          job.StartedAt,
		FinishedAt:         job.FinishedAt,
	}
}
