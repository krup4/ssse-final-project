package postgres

import (
	"context"

	"weather-accuracy/core-api/internal/domain"
)

func (r *RegionRepository) List(ctx context.Context) ([]domain.Region, error) {
	var models []RegionModel
	if err := r.db.WithContext(ctx).Order("name").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Region, 0, len(models))
	for _, model := range models {
		out = append(out, domain.Region{ID: model.ID, Name: model.Name})
	}
	return out, nil
}
