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
