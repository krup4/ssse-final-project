package postgres

import (
	"context"
	"strconv"
	"time"

	"weather-accuracy/core-api/internal/domain"
)

type stationErrorValue struct {
	ForecastValue float64
	ActualValue   float64
}

func (r *StationRepository) List(ctx context.Context, _ string, _ domain.StationStatus) ([]domain.Station, error) {
	var models []StationModel
	if err := r.db.WithContext(ctx).Where("is_active = true").Order("name").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Station, 0, len(models))
	for _, model := range models {
		station := domain.Station{
			ID:            strconv.Itoa(model.ID),
			Name:          model.Name,
			Lat:           model.Lat,
			Lon:           model.Lon,
			Status:        domain.StationOnline,
			ActiveSensors: 1,
		}
		var lastArchive time.Time
		_ = r.db.WithContext(ctx).Model(&ArchiveModel{}).Where("station_id = ?", model.ID).Select("max(dt)").Scan(&lastArchive).Error
		var lastForecast time.Time
		_ = r.db.WithContext(ctx).Model(&ForecastModel{}).Where("station_id = ?", model.ID).Select("max(date)").Scan(&lastForecast).Error
		if lastArchive.After(lastForecast) {
			station.LastTelemetryAt = lastArchive
		} else {
			station.LastTelemetryAt = lastForecast
		}
		var values []stationErrorValue
		_ = r.db.WithContext(ctx).Raw(`
			select f.value as forecast_value, a.value as actual_value
			from forecasts f
			join forecast_fields ff on ff.id = f.field_id
			join metrics m on m.forecast_field_id = ff.id
			join archive a on a.station_id = f.station_id and a.metric_id = m.id and a.dt = f.date
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
