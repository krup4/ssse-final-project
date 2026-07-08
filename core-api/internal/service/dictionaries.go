package service

import (
	"context"

	"weather-accuracy/core-api/internal/domain"
)

func (s *Service) ListRegions(ctx context.Context) ([]domain.Region, error) {
	return s.deps.Regions.List(ctx)
}

func (s *Service) ListForecastFields(ctx context.Context) ([]domain.ForecastField, error) {
	return s.deps.ForecastFields.ListWithArchiveData(ctx)
}

func (s *Service) ListMetrics(ctx context.Context) ([]domain.MetricDefinition, error) {
	return s.deps.ForecastFields.ListMetrics(ctx)
}

func (s *Service) ListStations(ctx context.Context, regionID string, status domain.StationStatus) ([]domain.Station, error) {
	return s.deps.Stations.List(ctx, regionID, status)
}

func (s *Service) CreateStation(ctx context.Context, input domain.StationInput, role domain.UserRole) (domain.Station, error) {
	if err := s.RequireRole(role, domain.RoleAdmin); err != nil {
		return domain.Station{}, err
	}
	return s.deps.Stations.Create(ctx, input)
}

func (s *Service) UpdateStation(ctx context.Context, id string, input domain.StationInput, role domain.UserRole) (domain.Station, error) {
	if err := s.RequireRole(role, domain.RoleAdmin); err != nil {
		return domain.Station{}, err
	}
	return s.deps.Stations.Update(ctx, id, input)
}
