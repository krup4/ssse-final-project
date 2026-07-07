package postgres

import (
	"context"
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm/clause"

	"weather-accuracy/core-api/internal/domain"
)

func (r *ActualWeatherRepository) SaveReading(ctx context.Context, reading domain.ActualWeatherReading) error {
	model := ActualWeatherReadingModel{
		ID:         reading.ID,
		StationID:  reading.StationID,
		ObservedAt: reading.ObservedAt,
		Source:     reading.Source,
		TraceID:    reading.TraceID,
		RawPayload: datatypes.JSON(reading.RawPayload),
	}
	if model.ID == "" {
		model.ID = fmt.Sprintf("%s-%s-%s", reading.StationID, reading.ObservedAt.Format(time.RFC3339), reading.TraceID)
	}
	for parameter, value := range reading.Values {
		setParameter(&model, parameter, value)
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"temperature_min", "temperature_max", "precipitation_total", "wind_speed", "wind_gust", "humidity", "pressure", "raw_payload", "created_at"}),
	}).Create(&model).Error
}
