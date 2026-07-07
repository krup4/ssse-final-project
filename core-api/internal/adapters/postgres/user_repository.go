package postgres

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"weather-accuracy/core-api/internal/domain"
)

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("email = ?", strings.ToLower(email)).First(&model).Error
	return toUser(model), mapError(err)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	return toUser(model), mapError(err)
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Order("email").Find(&models).Error; err != nil {
		return nil, err
	}
	out := make([]domain.User, 0, len(models))
	for _, model := range models {
		out = append(out, toUser(model))
	}
	return out, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id string, role domain.UserRole) (domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&model, "id = ?", id).Error; err != nil {
			return err
		}
		model.Role = string(role)
		return tx.Save(&model).Error
	})
	return toUser(model), mapError(err)
}

func (r *UserRepository) TouchLastSeen(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Update("last_seen", time.Now().UTC()).Error
}
