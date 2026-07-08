package postgres

import (
	"context"

	"weather-accuracy/core-api/internal/domain"
)

type stationErrorValue struct {
	ForecastValue float64
	ActualValue   float64
}

func (r *StationRepository) List(ctx context.Context, regionID string, status domain.StationStatus) ([]domain.Station, error) {
	query := r.db.WithContext(ctx).Model(&StationModel{})
	if regionID != "" && regionID != "all" {
		query = query.Where("region_id = ?", regionID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var models []StationModel
	if err := query.Order("name").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Station, 0, len(models))
	for _, model := range models {
		station := domain.Station{
			ID:              model.ID,
			Name:            model.Name,
			RegionID:        model.RegionID,
			Lat:             model.Lat,
			Lon:             model.Lon,
			Status:          domain.StationStatus(model.Status),
			ActiveSensors:   model.ActiveSensors,
			LastTelemetryAt: model.LastTelemetryAt,
		}
		var values []stationErrorValue
		_ = r.db.WithContext(ctx).Raw(`
			select f.value as forecast_value,
			       v.actual_value
			from forecast_reading_models f
			join actual_weather_reading_models a on a.station_id = f.station_id and a.observed_at = f.target_at
			join lateral (
				select unnest(array['temperature','wind_speed','humidity','pressure','precipitation']) metric,
				       unnest(array[a.temperature,a.wind_speed,a.humidity,a.pressure,a.precipitation_total]) actual_value
			) v on v.metric = f.metric and v.actual_value is not null
			where f.station_id = ?`, model.ID).Scan(&values).Error
		for _, value := range values {
			absoluteError, _ := calculateError(value.ForecastValue, value.ActualValue)
			if absoluteError > station.MaxError {
				station.MaxError = absoluteError
			}
		}
		out = append(out, station)
	}
	return out, nil
}
