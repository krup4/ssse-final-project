--liquibase formatted sql

--changeset core-api:002-demo-seed context:demo
INSERT INTO user_models (id, name, email, password_hash, role, last_seen, created_at, updated_at)
VALUES
    ('u-1', 'Ada Admin', 'admin@weather.local', '$2a$10$9pkz0gdrSix7E8KqfbgdhOF4NYC4ruRKPAQBX/UdKimaD4gNGCGwy', 'admin', now(), now(), now()),
    ('u-2', 'Alex Analyst', 'analyst@weather.local', '$2a$10$9pkz0gdrSix7E8KqfbgdhOF4NYC4ruRKPAQBX/UdKimaD4gNGCGwy', 'analyst', now(), now(), now()),
    ('u-3', 'Olga Operator', 'operator@weather.local', '$2a$10$9pkz0gdrSix7E8KqfbgdhOF4NYC4ruRKPAQBX/UdKimaD4gNGCGwy', 'operator', now(), now(), now())
ON CONFLICT DO NOTHING;

INSERT INTO region_models (id, name)
VALUES
    ('central', 'Central District'),
    ('north', 'Northern District'),
    ('south', 'Southern District')
ON CONFLICT DO NOTHING;

INSERT INTO station_models (id, name, region_id, lat, lon, status, active_sensors, last_telemetry_at)
VALUES
    ('st-004', 'Ryazan Field', 'central', 54.626, 39.735, 'online', 8, now() - interval '10 minutes'),
    ('st-006', 'Caspian Steppe', 'south', 46.349, 48.041, 'degraded', 5, now() - interval '35 minutes'),
    ('st-011', 'Murmansk Port', 'north', 68.958, 33.082, 'offline', 0, now() - interval '6 hours')
ON CONFLICT DO NOTHING;

INSERT INTO alert_models (id, title, source, severity, status, started_at, updated_at)
VALUES ('a-1', 'Kafka lag above SLO', 'Telemetry Receiver', 'warning', 'open', now() - interval '40 minutes', now() - interval '5 minutes')
ON CONFLICT DO NOTHING;

WITH station_list AS (
    SELECT unnest(array['st-004', 'st-006', 'st-011']) AS station_id
),
metric_list AS (
    SELECT *
    FROM (VALUES
        ('temperature', 0),
        ('wind_speed', 1),
        ('humidity', 2),
        ('pressure', 3),
        ('precipitation', 4)
    ) AS m(metric, metric_idx)
),
hours AS (
    SELECT generate_series(0, 35) AS h
),
points AS (
    SELECT
        s.station_id,
        m.metric,
        m.metric_idx,
        h.h,
        date_trunc('hour', now()) - (h.h || ' hours')::interval AS ts,
        10 + (m.metric_idx * 7) + (h.h % 5) AS forecast_value,
        10 + (m.metric_idx * 7) + (h.h % 5) + (((h.h + m.metric_idx) % 9) - 4) AS actual_value
    FROM station_list s
    CROSS JOIN metric_list m
    CROSS JOIN hours h
)
INSERT INTO forecast_reading_models (id, station_id, metric, value, forecast_at, target_at, source)
SELECT
    station_id || '-' || metric || '-' || to_char(ts, 'YYYYMMDDHH24'),
    station_id,
    metric,
    forecast_value,
    ts - interval '6 hours',
    ts,
    'demo-forecast-api'
FROM points
ON CONFLICT DO NOTHING;

WITH station_list AS (
    SELECT unnest(array['st-004', 'st-006', 'st-011']) AS station_id
),
metric_list AS (
    SELECT *
    FROM (VALUES
        ('temperature', 0),
        ('wind_speed', 1),
        ('humidity', 2),
        ('pressure', 3),
        ('precipitation', 4)
    ) AS m(metric, metric_idx)
),
hours AS (
    SELECT generate_series(0, 35) AS h
),
points AS (
    SELECT
        s.station_id,
        m.metric,
        m.metric_idx,
        h.h,
        date_trunc('hour', now()) - (h.h || ' hours')::interval AS ts,
        10 + (m.metric_idx * 7) + (h.h % 5) + (((h.h + m.metric_idx) % 9) - 4) AS actual_value
    FROM station_list s
    CROSS JOIN metric_list m
    CROSS JOIN hours h
)
INSERT INTO actual_weather_reading_models (
    id,
    station_id,
    observed_at,
    temperature_max,
    wind_speed,
    humidity,
    pressure,
    precipitation_total,
    source
)
SELECT
    station_id || '-actual-' || metric || '-' || to_char(ts, 'YYYYMMDDHH24'),
    station_id,
    ts,
    CASE WHEN metric = 'temperature' THEN actual_value END,
    CASE WHEN metric = 'wind_speed' THEN actual_value END,
    CASE WHEN metric = 'humidity' THEN actual_value END,
    CASE WHEN metric = 'pressure' THEN actual_value END,
    CASE WHEN metric = 'precipitation' THEN actual_value END,
    'demo-sensor'
FROM points
ON CONFLICT DO NOTHING;

--rollback DELETE FROM actual_weather_reading_models WHERE source = 'demo-sensor';
--rollback DELETE FROM forecast_reading_models WHERE source = 'demo-forecast-api';
--rollback DELETE FROM alert_models WHERE id = 'a-1';
--rollback DELETE FROM station_models WHERE id IN ('st-004', 'st-006', 'st-011');
--rollback DELETE FROM region_models WHERE id IN ('central', 'north', 'south');
--rollback DELETE FROM user_models WHERE id IN ('u-1', 'u-2', 'u-3');
