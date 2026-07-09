package postgres

import (
	"context"
	"math"
	"strconv"
	"time"

	"weather-accuracy/core-api/internal/domain"

	"gorm.io/gorm"
)

func (r *ActualWeatherRepository) SaveReading(ctx context.Context, reading domain.ActualWeatherReading) error {
	stationID, err := strconv.Atoi(reading.StationID)
	if err != nil {
		return err
	}
	observedAt := reading.ObservedAt.UTC().Truncate(time.Hour)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&StationModel{}, stationID).Error; err != nil {
			return err
		}
		for parameter, actualValue := range reading.Values {
			if err := saveParameterMetrics(ctx, tx, stationID, observedAt, parameter, actualValue); err != nil {
				return err
			}
		}
		return nil
	})
}

func saveParameterMetrics(ctx context.Context, tx *gorm.DB, stationID int, observedAt time.Time, parameter domain.WeatherParameter, actualValue float64) error {
	field := ForecastFieldModel{Name: string(parameter)}
	if err := tx.WithContext(ctx).Where("name = ?", field.Name).FirstOrCreate(&field).Error; err != nil {
		return err
	}

	var forecast ForecastModel
	if err := tx.WithContext(ctx).
		Where("station_id = ? AND field_id = ? AND date = ?", stationID, field.ID, observedAt).
		First(&forecast).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}

	absoluteError := math.Abs(forecast.Value - actualValue)
	values := map[string]float64{
		string(domain.MetricMAE):  absoluteError,
		string(domain.MetricRMSE): absoluteError,
		string(domain.MetricMSE):  absoluteError * absoluteError,
	}
	for metricName, metricValue := range values {
		metric := MetricModel{ForecastFieldID: field.ID, Name: metricName}
		if err := tx.WithContext(ctx).
			Where("forecast_field_id = ? AND name = ?", field.ID, metricName).
			FirstOrCreate(&metric).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Exec(`
			INSERT INTO archive (dt, station_id, metric_id, value)
			VALUES (?, ?, ?, ?)
			ON CONFLICT (station_id, metric_id, dt) DO UPDATE
			SET value = excluded.value`,
			observedAt,
			stationID,
			metric.ID,
			metricValue,
		).Error; err != nil {
			return err
		}
	}
	return tx.WithContext(ctx).
		Model(&ForecastModel{}).
		Where("station_id = ? AND field_id = ? AND date = ?", stationID, field.ID, observedAt).
		Update("is_archived", true).Error
}
