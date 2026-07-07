package weather

import (
    "context"
    "testing"
    "time"

    "actual-weather-feather/internal/models"
)

type fakeOMClient struct{
    meas *models.WeatherMeasurement
    err error
}

func (f *fakeOMClient) GetCurrentMeasurement(ctx context.Context, lat, lon float64) (*models.WeatherMeasurement, error) {
    return f.meas, f.err
}

type fakeRepo struct{
    saved *models.WeatherMeasurement
    err error
}

func (r *fakeRepo) Save(ctx context.Context, m *models.WeatherMeasurement) error {
    if r.err != nil { return r.err }
    r.saved = m
    return nil
}
func (r *fakeRepo) GetLatestByStation(ctx context.Context, stationID int64) (*models.WeatherMeasurement, error) { return nil, nil }

type fakeStationRepo struct{
    stations []models.Station
}
func (r *fakeStationRepo) GetAll(ctx context.Context) ([]models.Station, error) { return r.stations, nil }
func (r *fakeStationRepo) GetByID(ctx context.Context, id int64) (models.Station, error) { return models.Station{}, nil }
func (r *fakeStationRepo) Create(ctx context.Context, s *models.Station) error { return nil }
func (r *fakeStationRepo) Update(ctx context.Context, s *models.Station) error { return nil }
func (r *fakeStationRepo) Delete(ctx context.Context, id int64) error { return nil }

type fakePublisher struct{
    published any
}
func (p *fakePublisher) Publish(ctx context.Context, key string, value any) error { p.published = value; return nil }
func (p *fakePublisher) Close() error { return nil }

type fakeRedisClient struct{}

func (f *fakeRedisClient) Ping(ctx context.Context) error { return nil }
func (f *fakeRedisClient) Get(ctx context.Context, key string) (string, error) { return "", nil }
func (f *fakeRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error { return nil }
func (f *fakeRedisClient) Close() error { return nil }

// We reuse monitoring.Metrics counters in production; for test we pass nils where not used

func TestCollectForStation_SavesAndPublishes(t *testing.T) {
    now := time.Now().UTC()
    meas := &models.WeatherMeasurement{Timestamp: now, Temperature: ptrFloat(10.0)}

    omc := &fakeOMClient{meas: meas}
    repo := &fakeRepo{}
    srepo := &fakeStationRepo{}
    pub := &fakePublisher{}

    cache := &fakeRedisClient{}
    svc := NewService(omc, repo, srepo, pub, cache, nil)

    station := models.Station{ID: 1}
    if err := svc.CollectForStation(context.Background(), station); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if repo.saved == nil {
        t.Fatalf("expected saved measurement, got nil")
    }
    if pub.published == nil {
        t.Fatalf("expected published event, got nil")
    }
}

func ptrFloat(v float64) *float64 { return &v }
