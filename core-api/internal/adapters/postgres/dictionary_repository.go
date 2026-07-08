package postgres

import (
	"context"
	"strconv"

	"weather-accuracy/core-api/internal/domain"
)

func (r *RegionRepository) List(ctx context.Context) ([]domain.Region, error) {
	return []domain.Region{}, nil
}

func (r *ForecastFieldRepository) ListWithArchiveData(ctx context.Context) ([]domain.ForecastField, error) {
	var models []ForecastFieldModel
	err := r.db.WithContext(ctx).
		Model(&ForecastFieldModel{}).
		Joins("join metrics on metrics.forecast_field_id = forecast_fields.id").
		Joins("join archive on archive.metric_id = metrics.id").
		Group("forecast_fields.id, forecast_fields.name").
		Order("forecast_fields.name").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.ForecastField, 0, len(models))
	for _, model := range models {
		out = append(out, domain.ForecastField{ID: strconv.Itoa(model.ID), Name: model.Name})
	}
	return out, nil
}

func (r *ForecastFieldRepository) ListMetrics(ctx context.Context) ([]domain.MetricDefinition, error) {
	var models []MetricModel
	err := r.db.WithContext(ctx).
		Preload("ForecastField").
		Order("name").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.MetricDefinition, 0, len(models))
	for _, model := range models {
		out = append(out, domain.MetricDefinition{
			ID:              strconv.Itoa(model.ID),
			ForecastFieldID: strconv.Itoa(model.ForecastFieldID),
			ForecastField:   model.ForecastField.Name,
			Name:            model.Name,
		})
	}
	return out, nil
}
