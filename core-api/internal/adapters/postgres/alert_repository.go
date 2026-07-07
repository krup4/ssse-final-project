package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"weather-accuracy/core-api/internal/domain"
)

func (r *AlertRepository) List(ctx context.Context, status domain.AlertStatus, severity domain.AlertSeverity, source string) ([]domain.Alert, error) {
	query := r.db.WithContext(ctx).Model(&AlertModel{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}
	if source != "" {
		query = query.Where("source = ?", source)
	}
	var models []AlertModel
	if err := query.Order("started_at desc").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.Alert, 0, len(models))
	for _, model := range models {
		out = append(out, toAlert(model))
	}
	return out, nil
}

func (r *AlertRepository) UpdateStatus(ctx context.Context, id string, status domain.AlertStatus) (domain.Alert, error) {
	var model AlertModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&model, "id = ?", id).Error; err != nil {
			return err
		}
		model.Status = string(status)
		model.UpdatedAt = time.Now().UTC()
		return tx.Save(&model).Error
	})
	return toAlert(model), mapError(err)
}
