package service

import (
	"context"

	"weather-accuracy/core-api/internal/domain"
)

func (s *Service) ListAlerts(ctx context.Context, status domain.AlertStatus, severity domain.AlertSeverity, source string) ([]domain.Alert, error) {
	return s.deps.Alerts.List(ctx, status, severity, source)
}

func (s *Service) UpdateAlert(ctx context.Context, id string, status domain.AlertStatus, role domain.UserRole) (domain.Alert, error) {
	if err := s.RequireRole(role, domain.RoleAdmin, domain.RoleOperator); err != nil {
		return domain.Alert{}, err
	}
	return s.deps.Alerts.UpdateStatus(ctx, id, status)
}
