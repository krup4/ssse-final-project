package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Options struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectRetries  int
	ConnectBackoff  time.Duration
	PingTimeout     time.Duration
}

type Repositories struct {
	db            *gorm.DB
	Users         *UserRepository
	Regions       *RegionRepository
	Stations      *StationRepository
	Analytics     *AnalyticsRepository
	Alerts        *AlertRepository
	Backfills     *BackfillRepository
	ActualWeather *ActualWeatherRepository
}

func Open(ctx context.Context, options Options) (*gorm.DB, error) {
	if options.ConnectRetries <= 0 {
		options.ConnectRetries = 1
	}
	if options.ConnectBackoff <= 0 {
		options.ConnectBackoff = time.Second
	}
	if options.PingTimeout <= 0 {
		options.PingTimeout = 2 * time.Second
	}

	var lastErr error
	backoff := options.ConnectBackoff
	for attempt := 1; attempt <= options.ConnectRetries; attempt++ {
		db, err := openOnce(ctx, options)
		if err == nil {
			return db, nil
		}
		lastErr = err
		if attempt == options.ConnectRetries {
			break
		}
		if !sleep(ctx, backoff) {
			return nil, ctx.Err()
		}
		backoff = minDuration(backoff*2, 30*time.Second)
	}
	return nil, fmt.Errorf("postgres connection failed after %d attempts: %w", options.ConnectRetries, lastErr)
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		db:            db,
		Users:         &UserRepository{db: db},
		Regions:       &RegionRepository{db: db},
		Stations:      &StationRepository{db: db},
		Analytics:     &AnalyticsRepository{db: db},
		Alerts:        &AlertRepository{db: db},
		Backfills:     &BackfillRepository{db: db},
		ActualWeather: &ActualWeatherRepository{db: db},
	}
}

func (r *Repositories) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func openOnce(ctx context.Context, options Options) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(options.URL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	configurePool(sqlDB, options)
	pingCtx, cancel := context.WithTimeout(ctx, options.PingTimeout)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func configurePool(db *sql.DB, options Options) {
	if options.MaxOpenConns > 0 {
		db.SetMaxOpenConns(options.MaxOpenConns)
	}
	if options.MaxIdleConns > 0 {
		db.SetMaxIdleConns(options.MaxIdleConns)
	}
	if options.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(options.ConnMaxLifetime)
	}
	if options.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(options.ConnMaxIdleTime)
	}
}

func (r *Repositories) Migrate(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(
		&UserModel{},
		&RegionModel{},
		&StationModel{},
		&ForecastReadingModel{},
		&ActualWeatherReadingModel{},
		&AlertModel{},
		&BackfillJobModel{},
	)
}

func (r *Repositories) SeedDemoData(ctx context.Context) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	users := []UserModel{
		{ID: "u-1", Name: "Ada Admin", Email: "admin@weather.local", PasswordHash: string(hash), Role: "admin", LastSeen: now},
		{ID: "u-2", Name: "Alex Analyst", Email: "analyst@weather.local", PasswordHash: string(hash), Role: "analyst", LastSeen: now},
		{ID: "u-3", Name: "Olga Operator", Email: "operator@weather.local", PasswordHash: string(hash), Role: "operator", LastSeen: now},
	}
	regions := []RegionModel{
		{ID: "central", Name: "Central District"},
		{ID: "north", Name: "Northern District"},
		{ID: "south", Name: "Southern District"},
	}
	stations := []StationModel{
		{ID: "st-004", Name: "Ryazan Field", RegionID: "central", Lat: 54.626, Lon: 39.735, Status: "online", ActiveSensors: 8, LastTelemetryAt: now.Add(-10 * time.Minute)},
		{ID: "st-006", Name: "Caspian Steppe", RegionID: "south", Lat: 46.349, Lon: 48.041, Status: "degraded", ActiveSensors: 5, LastTelemetryAt: now.Add(-35 * time.Minute)},
		{ID: "st-011", Name: "Murmansk Port", RegionID: "north", Lat: 68.958, Lon: 33.082, Status: "offline", ActiveSensors: 0, LastTelemetryAt: now.Add(-6 * time.Hour)},
	}
	alerts := []AlertModel{
		{ID: "a-1", Title: "Kafka lag above SLO", Source: "Telemetry Receiver", Severity: "warning", Status: "open", StartedAt: now.Add(-40 * time.Minute), UpdatedAt: now.Add(-5 * time.Minute)},
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&users).Error; err != nil {
			return err
		}
		if err := tx.Create(&regions).Error; err != nil {
			return err
		}
		if err := tx.Create(&stations).Error; err != nil {
			return err
		}
		if err := tx.Create(&alerts).Error; err != nil {
			return err
		}
		return seedReadings(tx, now)
	})
}

func seedReadings(tx *gorm.DB, now time.Time) error {
	metrics := []string{"temperature", "wind_speed", "humidity", "pressure", "precipitation"}
	stations := []string{"st-004", "st-006", "st-011"}
	for h := 0; h < 36; h++ {
		ts := now.Truncate(time.Hour).Add(time.Duration(-h) * time.Hour)
		for _, station := range stations {
			for i, metric := range metrics {
				forecast := 10 + float64(i*7) + float64(h%5)
				actual := forecast + float64((h+i)%9-4)
				fr := ForecastReadingModel{
					ID:         station + "-" + metric + "-" + ts.Format("2006010215"),
					StationID:  station,
					Metric:     metric,
					Value:      forecast,
					ForecastAt: ts.Add(-6 * time.Hour),
					TargetAt:   ts,
					Source:     "demo-forecast-api",
				}
				ar := ActualWeatherReadingModel{
					ID:         station + "-actual-" + metric + "-" + ts.Format("2006010215"),
					StationID:  station,
					ObservedAt: ts,
					Source:     "demo-sensor",
				}
				setActualMetric(&ar, metric, actual)
				if err := tx.Create(&fr).Error; err != nil && !isDuplicate(err) {
					return err
				}
				if err := tx.Create(&ar).Error; err != nil && !isDuplicate(err) {
					return err
				}
			}
		}
	}
	return nil
}

func isDuplicate(err error) bool {
	return err != nil && errors.Is(err, gorm.ErrDuplicatedKey)
}

func sleep(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
