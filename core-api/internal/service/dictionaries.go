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

func (s *Service) ListStations(ctx context.Context, regionID string, status domain.StationStatus) ([]domain.Station, error) {
	return s.deps.Stations.List(ctx, regionID, status)
}
