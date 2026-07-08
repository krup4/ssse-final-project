package postgres

import (
	"errors"
	"strconv"

	"gorm.io/gorm"

	"weather-accuracy/core-api/internal/domain"
)

func toUser(model UserModel) domain.User {
	return domain.User{
		ID:           strconv.Itoa(model.ID),
		Login:        model.Login,
		Name:         model.Name,
		Email:        model.Email,
		Role:         domain.UserRole(model.Role.Name),
		IsActive:     model.IsActive,
		LastSeen:     model.LastSeen,
		PasswordHash: model.PasswordHash,
	}
}

func toAlert(model AlertModel) domain.Alert {
	return domain.Alert{
		ID:        model.ID,
		Title:     model.Title,
		Source:    model.Source,
		Severity:  domain.AlertSeverity(model.Severity),
		Status:    domain.AlertStatus(model.Status),
		StartedAt: model.StartedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func toBackfill(model BackfillJobModel) domain.BackfillJob {
	return domain.BackfillJob{
		ID:                 model.ID,
		Status:             domain.BackfillStatus(model.Status),
		CalculationVersion: model.CalculationVersion,
		ProgressPct:        model.ProgressPct,
		ProcessedRows:      model.ProcessedRows,
		FailedRows:         model.FailedRows,
		CreatedAt:          model.CreatedAt,
		StartedAt:          model.StartedAt,
		FinishedAt:         model.FinishedAt,
	}
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	return err
}
