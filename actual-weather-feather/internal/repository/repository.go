package repository

import (
    "context"

    "actual-weather-feather/internal/models"
    "errors"
)

type StationRepository interface {
    GetAll(ctx context.Context) ([]models.Station, error)
    GetByID(ctx context.Context, id int64) (models.Station, error)
    Create(ctx context.Context, s *models.Station) error
    Update(ctx context.Context, s *models.Station) error
    Delete(ctx context.Context, id int64) error
}

type WeatherRepository interface {
    Save(ctx context.Context, m *models.WeatherMeasurement) error
    GetLatestByStation(ctx context.Context, stationID int64) (*models.WeatherMeasurement, error)
}

var ErrNotFound = errors.New("not found")
