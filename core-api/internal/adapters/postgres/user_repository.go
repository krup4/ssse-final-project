package postgres

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"weather-accuracy/core-api/internal/domain"
)

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Preload("Role").Where("login = ?", strings.ToLower(login)).First(&model).Error
	return toUser(model), mapError(err)
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Preload("Role").First(&model, "id = ?", id).Error
	return toUser(model), mapError(err)
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	var models []UserModel
	if err := r.db.WithContext(ctx).Preload("Role").Order("login").Find(&models).Error; err != nil {
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
		var roleModel RoleModel
		if err := tx.First(&roleModel, "name = ?", string(role)).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Role").First(&model, "id = ?", id).Error; err != nil {
			return err
		}
		model.RoleID = roleModel.ID
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
		model.Role = roleModel
		return nil
	})
	return toUser(model), mapError(err)
}

func (r *UserRepository) TouchLastSeen(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Update("last_seen", time.Now().UTC()).Error
}
