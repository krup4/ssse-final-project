package postgres

import (
	"context"
	"strconv"

	"weather-accuracy/core-api/internal/domain"
)

func (r *ActualWeatherRepository) SaveReading(ctx context.Context, reading domain.ActualWeatherReading) error {
	stationID, err := strconv.Atoi(reading.StationID)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).First(&StationModel{}, stationID).Error
}
