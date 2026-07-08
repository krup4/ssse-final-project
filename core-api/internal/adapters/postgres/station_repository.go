package postgres

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"weather-accuracy/core-api/internal/domain"
)

type stationErrorValue struct {
	ForecastValue float64
	ActualValue   float64
}

func (r *StationRepository) List(ctx context.Context, _ string, status domain.StationStatus) ([]domain.Station, error) {
	var models []StationModel
	query := r.db.WithContext(ctx).Order("name")
	switch status {
	case domain.StationOnline:
		query = query.Where("is_active = true")
	case domain.StationOffline:
		query = query.Where("is_active = false")
	}
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Station, 0, len(models))
	for _, model := range models {
		station := r.toDomainStation(ctx, model)
		out = append(out, station)
	}
	return out, nil
}

func (r *StationRepository) Create(ctx context.Context, input domain.StationInput) (domain.Station, error) {
	model := StationModel{
		Name:     strings.TrimSpace(input.Name),
		Lat:      input.Lat,
		Lon:      input.Lon,
		IsActive: input.IsActive,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return domain.Station{}, err
	}
	return r.toDomainStation(ctx, model), nil
}

func (r *StationRepository) Update(ctx context.Context, id string, input domain.StationInput) (domain.Station, error) {
	stationID, err := strconv.Atoi(id)
	if err != nil {
		return domain.Station{}, domain.ErrNotFound
	}
	var model StationModel
	if err := r.db.WithContext(ctx).First(&model, stationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return domain.Station{}, domain.ErrNotFound
		}
		return domain.Station{}, err
	}
	model.Name = strings.TrimSpace(input.Name)
	model.Lat = input.Lat
	model.Lon = input.Lon
	model.IsActive = input.IsActive
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return domain.Station{}, err
	}
	return r.toDomainStation(ctx, model), nil
}

func (r *StationRepository) toDomainStation(ctx context.Context, model StationModel) domain.Station {
	status := domain.StationOnline
	activeSensors := 1
	if !model.IsActive {
		status = domain.StationOffline
		activeSensors = 0
	}
	station := domain.Station{
		ID:            strconv.Itoa(model.ID),
		Name:          model.Name,
		Lat:           model.Lat,
		Lon:           model.Lon,
		IsActive:      model.IsActive,
		Status:        status,
		ActiveSensors: activeSensors,
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
	return station
}
