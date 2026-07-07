package repository

import (
    "context"
    "fmt"

    "actual-weather-feather/internal/models"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type PgxRepository struct {
    db *pgxpool.Pool
}

func NewPgxRepository(db *pgxpool.Pool) *PgxRepository {
    return &PgxRepository{db: db}
}

func (p *PgxRepository) GetAll(ctx context.Context) ([]models.Station, error) {
    rows, err := p.db.Query(ctx, "SELECT id, name, latitude, longitude, region, active FROM stations WHERE active = true")
    if err != nil {
        return nil, fmt.Errorf("query stations: %w", err)
    }
    defer rows.Close()

    var res []models.Station
    for rows.Next() {
        var s models.Station
        if err := rows.Scan(&s.ID, &s.Name, &s.Latitude, &s.Longitude, &s.Region, &s.Active); err != nil {
            return nil, fmt.Errorf("scan station: %w", err)
        }
        res = append(res, s)
    }
    return res, nil
}

// Basic stubs for other methods
func (p *PgxRepository) GetByID(ctx context.Context, id int64) (models.Station, error) {
    var s models.Station
    row := p.db.QueryRow(ctx, "SELECT id, name, latitude, longitude, region, active FROM stations WHERE id = $1", id)
    if err := row.Scan(&s.ID, &s.Name, &s.Latitude, &s.Longitude, &s.Region, &s.Active); err != nil {
        if err == pgx.ErrNoRows {
            return s, ErrNotFound
        }
        return s, err
    }
    return s, nil
}

func (p *PgxRepository) Create(ctx context.Context, s *models.Station) error {
    row := p.db.QueryRow(ctx, "INSERT INTO stations(name, latitude, longitude, region, active) VALUES($1,$2,$3,$4,$5) RETURNING id", s.Name, s.Latitude, s.Longitude, s.Region, s.Active)
    if err := row.Scan(&s.ID); err != nil {
        return err
    }
    return nil
}

func (p *PgxRepository) Update(ctx context.Context, s *models.Station) error {
    cmd, err := p.db.Exec(ctx, "UPDATE stations SET name=$1, latitude=$2, longitude=$3, region=$4, active=$5 WHERE id=$6", s.Name, s.Latitude, s.Longitude, s.Region, s.Active, s.ID)
    if err != nil {
        return err
    }
    if cmd.RowsAffected() == 0 {
        return ErrNotFound
    }
    return nil
}

func (p *PgxRepository) Delete(ctx context.Context, id int64) error {
    cmd, err := p.db.Exec(ctx, "DELETE FROM stations WHERE id=$1", id)
    if err != nil {
        return err
    }
    if cmd.RowsAffected() == 0 {
        return ErrNotFound
    }
    return nil
}

// Weather repository minimal implementation
func (p *PgxRepository) Save(ctx context.Context, m *models.WeatherMeasurement) error {
    _, err := p.db.Exec(ctx, `INSERT INTO weather_measurements(station_id, timestamp, temperature, humidity, pressure, wind_speed, wind_direction, precipitation, created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
        m.StationID, m.Timestamp, m.Temperature, m.Humidity, m.Pressure, m.WindSpeed, m.WindDirection, m.Precipitation, m.CreatedAt)
    return err
}

func (p *PgxRepository) GetLatestByStation(ctx context.Context, stationID int64) (*models.WeatherMeasurement, error) {
    var m models.WeatherMeasurement
    row := p.db.QueryRow(ctx, "SELECT id, station_id, timestamp, temperature, humidity, pressure, wind_speed, wind_direction, precipitation, created_at FROM weather_measurements WHERE station_id=$1 ORDER BY timestamp DESC LIMIT 1", stationID)
    if err := row.Scan(&m.ID, &m.StationID, &m.Timestamp, &m.Temperature, &m.Humidity, &m.Pressure, &m.WindSpeed, &m.WindDirection, &m.Precipitation, &m.CreatedAt); err != nil {
        return nil, err
    }
    return &m, nil
}
