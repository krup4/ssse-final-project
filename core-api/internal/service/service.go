package service

import (
	"time"

	"weather-accuracy/core-api/internal/domain"
)

type Dependencies struct {
	Users             domain.UserRepository
	Regions           domain.RegionRepository
	ForecastFields    domain.ForecastFieldRepository
	Stations          domain.StationRepository
	Analytics         domain.AnalyticsRepository
	Alerts            domain.AlertRepository
	Backfills         domain.BackfillRepository
	BackfillPublisher domain.BackfillPublisher
	JWTSecret         string
	TokenTTL          time.Duration
}

type Service struct {
	deps Dependencies
}

func New(deps Dependencies) *Service {
	return &Service{deps: deps}
}

func (s *Service) RequireRole(actual domain.UserRole, allowed ...domain.UserRole) error {
	for _, role := range allowed {
		if actual == role {
			return nil
		}
	}
	return domain.ErrForbidden
}
