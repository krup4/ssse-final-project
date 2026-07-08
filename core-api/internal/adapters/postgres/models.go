package postgres

import (
	"time"

	"gorm.io/datatypes"
)

type UserModel struct {
	ID           string    `gorm:"primaryKey;size:64"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"not null;index"`
	LastSeen     time.Time `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RegionModel struct {
	ID   string `gorm:"primaryKey;size:64"`
	Name string `gorm:"not null"`
}

type StationModel struct {
	ID              string    `gorm:"primaryKey;size:64"`
	Name            string    `gorm:"not null"`
	RegionID        string    `gorm:"not null;index"`
	Lat             float64   `gorm:"not null"`
	Lon             float64   `gorm:"not null"`
	Status          string    `gorm:"not null;index"`
	ActiveSensors   int       `gorm:"not null"`
	LastTelemetryAt time.Time `gorm:"not null;index"`
	Region          RegionModel
}

type ForecastReadingModel struct {
	ID         string         `gorm:"primaryKey;size:80"`
	StationID  string         `gorm:"not null;index:idx_forecast_match,priority:1"`
	Metric     string         `gorm:"not null;index:idx_forecast_match,priority:2"`
	Value      float64        `gorm:"not null"`
	ForecastAt time.Time      `gorm:"not null;index"`
	TargetAt   time.Time      `gorm:"not null;index:idx_forecast_match,priority:3"`
	Source     string         `gorm:"not null"`
	RawPayload datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt  time.Time
}

type ActualWeatherReadingModel struct {
	ID                 string    `gorm:"primaryKey;size:120"`
	StationID          string    `gorm:"not null;index:idx_actual_station_observed,priority:1"`
	ObservedAt         time.Time `gorm:"not null;index:idx_actual_station_observed,priority:2"`
	Temperature        *float64
	PrecipitationTotal *float64
	WindSpeed          *float64
	WindGust           *float64
	Humidity           *float64
	Pressure           *float64
	Source             string         `gorm:"not null"`
	TraceID            string         `gorm:"index"`
	RawPayload         datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt          time.Time
}

type AlertModel struct {
	ID        string    `gorm:"primaryKey;size:80"`
	Title     string    `gorm:"not null"`
	Source    string    `gorm:"not null;index"`
	Severity  string    `gorm:"not null;index"`
	Status    string    `gorm:"not null;index"`
	StartedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type BackfillJobModel struct {
	ID                 string    `gorm:"primaryKey;size:80"`
	Status             string    `gorm:"not null;index"`
	DateFrom           time.Time `gorm:"not null;index"`
	DateTo             time.Time `gorm:"not null;index"`
	RegionID           string    `gorm:"index"`
	StationID          string    `gorm:"index"`
	Metric             string    `gorm:"index"`
	CalculationVersion string    `gorm:"not null;index"`
	ProgressPct        float64   `gorm:"not null"`
	ProcessedRows      int64     `gorm:"not null"`
	FailedRows         int64     `gorm:"not null"`
	CreatedAt          time.Time `gorm:"not null"`
	StartedAt          *time.Time
	FinishedAt         *time.Time
}
