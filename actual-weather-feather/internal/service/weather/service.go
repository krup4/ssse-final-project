package weather

import (
    "context"
    "fmt"
    "time"

    om "actual-weather-feather/internal/client/openmeteo"
    "actual-weather-feather/internal/models"
    "actual-weather-feather/internal/repository"
    kfk "actual-weather-feather/internal/kafka"
    "actual-weather-feather/internal/monitoring"
    "actual-weather-feather/internal/redis"
    redisclient "github.com/redis/go-redis/v9"
)

type Service struct {
    omClient    om.API
    repo        repository.WeatherRepository
    stationRepo repository.StationRepository
    publisher   kfk.Publisher
    cache       redis.Client
    metrics     *monitoring.Metrics
}

func NewService(omc om.API, repo repository.WeatherRepository, srepo repository.StationRepository, pub kfk.Publisher, cache redis.Client, m *monitoring.Metrics) *Service {
    return &Service{omClient: omc, repo: repo, stationRepo: srepo, publisher: pub, cache: cache, metrics: m}
}

func (s *Service) CollectForStation(ctx context.Context, station models.Station) error {
    // get raw data
    // get structured measurement from client (includes retry)
    meas, err := s.omClient.GetCurrentMeasurement(ctx, station.Latitude, station.Longitude)
    if err != nil {
        if s.metrics != nil {
            s.metrics.ErrorsTotal.Inc()
        }
        return err
    }
    meas.StationID = station.ID

    // Basic validation: require timestamp
    if meas.Timestamp.IsZero() {
        if s.metrics != nil {
            s.metrics.ErrorsTotal.Inc()
        }
        return fmt.Errorf("empty timestamp in measurement")
    }

    if s.cache != nil {
        key := fmt.Sprintf("station:%d:last_timestamp", station.ID)
        prev, err := s.cache.Get(ctx, key)
        if err == nil {
            if prevTS, perr := time.Parse(time.RFC3339, prev); perr == nil {
                if !meas.Timestamp.After(prevTS) {
                    return nil
                }
            }
        } else if err != redisclient.Nil {
            if s.metrics != nil {
                s.metrics.ErrorsTotal.Inc()
            }
            return fmt.Errorf("redis cache error: %w", err)
        }
    }

    if err := s.repo.Save(ctx, meas); err != nil {
        if s.metrics != nil {
            s.metrics.ErrorsTotal.Inc()
        }
        return err
    }

    if s.cache != nil {
        key := fmt.Sprintf("station:%d:last_timestamp", station.ID)
        _ = s.cache.Set(ctx, key, meas.Timestamp.Format(time.RFC3339), 2*time.Hour)
    }

    if s.publisher == nil {
        if s.metrics != nil {
            s.metrics.ErrorsTotal.Inc()
        }
        return fmt.Errorf("kafka publisher is not configured")
    }

    if err := s.publisher.Publish(ctx, "", meas); err != nil {
        if s.metrics != nil {
            s.metrics.ErrorsTotal.Inc()
        }
        return err
    }
    if s.metrics != nil {
        s.metrics.WeatherSavedTotal.Inc()
    }
    return nil
}
