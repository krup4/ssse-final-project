package postgres

import (
	"time"
)

type RoleModel struct {
	ID   int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"uniqueIndex;not null"`
}

func (RoleModel) TableName() string { return "roles" }

type UserModel struct {
	ID           int       `gorm:"primaryKey;autoIncrement"`
	Login        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	RoleID       int       `gorm:"not null;index"`
	IsActive     bool      `gorm:"not null;default:true"`
	Name         string    `gorm:"not null"`
	Email        string    `gorm:"not null"`
	LastSeen     time.Time `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Role         RoleModel `gorm:"foreignKey:RoleID"`
}

func (UserModel) TableName() string { return "users" }

type ForecastFieldModel struct {
	ID   int    `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"uniqueIndex;not null"`
}

func (ForecastFieldModel) TableName() string { return "forecast_fields" }

type StationModel struct {
	ID       int     `gorm:"primaryKey;autoIncrement"`
	Name     string  `gorm:"not null"`
	Lon      float64 `gorm:"not null"`
	Lat      float64 `gorm:"not null"`
	IsActive bool    `gorm:"not null;default:true;index"`
}

func (StationModel) TableName() string { return "stations" }

type ForecastModel struct {
	ID         int                `gorm:"primaryKey;autoIncrement"`
	Date       time.Time          `gorm:"not null;index:idx_forecasts_match,priority:3"`
	FieldID    int                `gorm:"not null;index:idx_forecasts_match,priority:2"`
	Value      float64            `gorm:"not null"`
	Interval   string             `gorm:"not null"`
	StationID  int                `gorm:"not null;index:idx_forecasts_match,priority:1"`
	IsArchived bool               `gorm:"not null;default:false;index"`
	Field      ForecastFieldModel `gorm:"foreignKey:FieldID"`
	Station    StationModel       `gorm:"foreignKey:StationID"`
}

func (ForecastModel) TableName() string { return "forecasts" }

type MetricModel struct {
	ID              int                `gorm:"primaryKey;autoIncrement"`
	ForecastFieldID int                `gorm:"not null;index;uniqueIndex:idx_metrics_field_name,priority:1"`
	Name            string             `gorm:"not null;index;uniqueIndex:idx_metrics_field_name,priority:2"`
	ForecastField   ForecastFieldModel `gorm:"foreignKey:ForecastFieldID"`
}

func (MetricModel) TableName() string { return "metrics" }

type ArchiveModel struct {
	ID        int          `gorm:"primaryKey;autoIncrement"`
	Dt        time.Time    `gorm:"not null;index:idx_archive_match,priority:3;uniqueIndex:idx_archive_unique_point,priority:3"`
	StationID int          `gorm:"not null;index:idx_archive_match,priority:1;uniqueIndex:idx_archive_unique_point,priority:1"`
	MetricID  int          `gorm:"not null;index:idx_archive_match,priority:2;uniqueIndex:idx_archive_unique_point,priority:2"`
	Value     float64      `gorm:"not null"`
	Station   StationModel `gorm:"foreignKey:StationID"`
	Metric    MetricModel  `gorm:"foreignKey:MetricID"`
}

func (ArchiveModel) TableName() string { return "archive" }

type AlertModel struct {
	ID        string    `gorm:"primaryKey;size:80"`
	Title     string    `gorm:"not null"`
	Source    string    `gorm:"not null;index"`
	Severity  string    `gorm:"not null;index"`
	Status    string    `gorm:"not null;index"`
	StartedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (AlertModel) TableName() string { return "alerts" }

type BackfillJobModel struct {
	ID                 string    `gorm:"primaryKey;size:80"`
	Status             string    `gorm:"not null;index"`
	DateFrom           time.Time `gorm:"not null;index"`
	DateTo             time.Time `gorm:"not null;index"`
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

func (BackfillJobModel) TableName() string { return "backfill_jobs" }
