package postgres

import (
	"context"
	"strconv"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"weather-accuracy/core-api/internal/domain"
)

func (r *ActualWeatherRepository) SaveReading(ctx context.Context, reading domain.ActualWeatherReading) error {
	stationID, err := strconv.Atoi(reading.StationID)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for parameter, value := range reading.Values {
			metricName := metricNameForParameter(parameter)
			var metric MetricModel
			err := tx.
				Joins("join forecast_fields on forecast_fields.id = metrics.forecast_field_id").
				Where("metrics.name = ? or forecast_fields.name = ?", metricName, metricName).
				First(&metric).Error
			if err != nil {
				return err
			}
			model := ArchiveModel{
				StationID: stationID,
				MetricID:  metric.ID,
				Dt:        reading.ObservedAt,
				Value:     value,
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "station_id"}, {Name: "metric_id"}, {Name: "dt"}},
				DoUpdates: clause.AssignmentColumns([]string{"value"}),
			}).Create(&model).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func metricNameForParameter(parameter domain.WeatherParameter) string {
	switch parameter {
	case domain.ParameterTemperature:
		return "temperature"
	case domain.ParameterPrecipitation:
		return "precipitation"
	default:
		return string(parameter)
	}
}
