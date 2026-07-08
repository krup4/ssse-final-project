package service

import (
	"context"

	"weather-accuracy/core-api/internal/domain"
)

func (s *Service) CreateBackfill(ctx context.Context, request domain.BackfillRequest, role domain.UserRole) (domain.BackfillJob, error) {
	if err := s.RequireRole(role, domain.RoleAdmin, domain.RoleAnalyst); err != nil {
		return domain.BackfillJob{}, err
	}
	conflict, err := s.deps.Backfills.HasActiveConflict(ctx, request)
	if err != nil {
		return domain.BackfillJob{}, err
	}
	if conflict {
		return domain.BackfillJob{}, domain.ErrConflict
	}
	job, err := s.deps.Backfills.Create(ctx, request)
	if err != nil {
		return domain.BackfillJob{}, err
	}
	if s.deps.BackfillPublisher != nil {
		if err := s.deps.BackfillPublisher.PublishBackfillRequested(ctx, job, request); err != nil {
			return domain.BackfillJob{}, err
		}
	}
	return job, nil
}

func (s *Service) GetBackfill(ctx context.Context, id string) (domain.BackfillJob, error) {
	return s.deps.Backfills.Get(ctx, id)
}
