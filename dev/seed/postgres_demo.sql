\set ON_ERROR_STOP on

BEGIN;

DELETE FROM archive WHERE station_id BETWEEN 101 AND 105;
DELETE FROM forecasts WHERE station_id BETWEEN 101 AND 105;
DELETE FROM alerts WHERE id LIKE 'dev-%';
DELETE FROM backfill_jobs WHERE id LIKE 'dev-%';
DELETE FROM stations WHERE id BETWEEN 101 AND 105;

INSERT INTO roles (name)
VALUES ('admin'), ('analyst'), ('operator'), ('viewer')
ON CONFLICT (name) DO NOTHING;

INSERT INTO users (login, password_hash, role_id, is_active, name, email, last_seen, created_at, updated_at)
SELECT rows.login, rows.password_hash, roles.id, true, rows.name, rows.email, now(), now(), now()
FROM (VALUES
    ('admin', 'Ada Admin', 'admin@weather.local', '$2a$10$9pkz0gdrSix7E8KqfbgdhOF4NYC4ruRKPAQBX/UdKimaD4gNGCGwy', 'admin'),
    ('analyst', 'Alex Analyst', 'analyst@weather.local', '$2a$10$9pkz0gdrSix7E8KqfbgdhOF4NYC4ruRKPAQBX/UdKimaD4gNGCGwy', 'analyst'),
    ('operator', 'Olga Operator', 'operator@weather.local', '$2a$10$9pkz0gdrSix7E8KqfbgdhOF4NYC4ruRKPAQBX/UdKimaD4gNGCGwy', 'operator'),
    ('viewer', 'Vera Viewer', 'viewer@weather.local', '$2a$10$9pkz0gdrSix7E8KqfbgdhOF4NYC4ruRKPAQBX/UdKimaD4gNGCGwy', 'viewer')
) AS rows(login, name, email, password_hash, role_name)
JOIN roles ON roles.name = rows.role_name
ON CONFLICT (login) DO UPDATE
SET name = EXCLUDED.name,
    email = EXCLUDED.email,
    role_id = EXCLUDED.role_id,
    is_active = true,
    updated_at = now();

INSERT INTO forecast_fields (name)
VALUES ('temperature'), ('wind_speed'), ('humidity'), ('pressure'), ('precipitation'), ('wind_gust')
ON CONFLICT (name) DO NOTHING;

INSERT INTO metrics (forecast_field_id, name)
SELECT id, name
FROM forecast_fields
WHERE NOT EXISTS (
    SELECT 1
    FROM metrics
    WHERE metrics.forecast_field_id = forecast_fields.id
      AND metrics.name = forecast_fields.name
);

INSERT INTO stations (id, name, lon, lat, is_active)
VALUES
    (101, 'Dev Ryazan Field', 39.735, 54.626, true),
    (102, 'Dev Caspian Steppe', 48.041, 46.349, true),
    (103, 'Dev Murmansk Port', 33.082, 68.958, true),
    (104, 'Dev Baikal North', 108.165, 53.558, true),
    (105, 'Dev Inactive Station', 37.617, 55.755, false)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    lon = EXCLUDED.lon,
    lat = EXCLUDED.lat,
    is_active = EXCLUDED.is_active;

SELECT setval(pg_get_serial_sequence('stations', 'id'), greatest((SELECT max(id) FROM stations), 105));

WITH station_list AS (
    SELECT id AS station_id, row_number() OVER (ORDER BY id) - 1 AS station_idx
    FROM stations
    WHERE id BETWEEN 101 AND 105
),
field_list AS (
    SELECT id AS field_id, name, row_number() OVER (ORDER BY id) - 1 AS metric_idx
    FROM forecast_fields
    WHERE name IN ('temperature', 'wind_speed', 'humidity', 'pressure', 'precipitation')
),
hours AS (
    SELECT generate_series(0, 719) AS h
),
points AS (
    SELECT
        s.station_id,
        s.station_idx,
        f.field_id,
        f.name,
        f.metric_idx,
        h.h,
        date_trunc('hour', now()) - (h.h || ' hours')::interval AS ts,
        12 + (s.station_idx * 2) + (f.metric_idx * 6) + (h.h % 6) AS forecast_value,
        12 + (s.station_idx * 2) + (f.metric_idx * 6) + (h.h % 6)
          + CASE
              WHEN s.station_id = 101 AND f.name = 'wind_speed' AND h.h BETWEEN 0 AND 12 THEN 18
              WHEN s.station_id = 102 AND f.name = 'pressure' AND h.h BETWEEN 72 AND 120 THEN -14
              WHEN s.station_id = 103 AND f.name = 'temperature' AND h.h BETWEEN 168 AND 240 THEN 11
              WHEN s.station_id = 104 AND f.name = 'precipitation' AND h.h BETWEEN 360 AND 432 THEN 16
              ELSE ((h.h + f.metric_idx + s.station_idx) % 9) - 4
            END AS actual_value
    FROM station_list s
    CROSS JOIN field_list f
    CROSS JOIN hours h
)
INSERT INTO forecasts (date, field_id, value, interval, station_id, is_archived)
SELECT ts, field_id, forecast_value, '6h', station_id, h > 8
FROM points;

WITH station_list AS (
    SELECT id AS station_id, row_number() OVER (ORDER BY id) - 1 AS station_idx
    FROM stations
    WHERE id BETWEEN 101 AND 105
),
metric_list AS (
    SELECT m.id AS metric_id, ff.name, row_number() OVER (ORDER BY m.id) - 1 AS metric_idx
    FROM metrics m
    JOIN forecast_fields ff ON ff.id = m.forecast_field_id
    WHERE ff.name IN ('temperature', 'wind_speed', 'humidity', 'pressure', 'precipitation')
),
hours AS (
    SELECT generate_series(0, 719) AS h
),
points AS (
    SELECT
        s.station_id,
        s.station_idx,
        m.metric_id,
        m.name,
        m.metric_idx,
        h.h,
        date_trunc('hour', now()) - (h.h || ' hours')::interval AS ts,
        12 + (s.station_idx * 2) + (m.metric_idx * 6) + (h.h % 6)
          + CASE
              WHEN s.station_id = 101 AND m.name = 'wind_speed' AND h.h BETWEEN 0 AND 12 THEN 18
              WHEN s.station_id = 102 AND m.name = 'pressure' AND h.h BETWEEN 72 AND 120 THEN -14
              WHEN s.station_id = 103 AND m.name = 'temperature' AND h.h BETWEEN 168 AND 240 THEN 11
              WHEN s.station_id = 104 AND m.name = 'precipitation' AND h.h BETWEEN 360 AND 432 THEN 16
              ELSE ((h.h + m.metric_idx + s.station_idx) % 9) - 4
            END AS actual_value
    FROM station_list s
    CROSS JOIN metric_list m
    CROSS JOIN hours h
)
INSERT INTO archive (dt, station_id, metric_id, value)
SELECT ts, station_id, metric_id, actual_value
FROM points
ON CONFLICT (station_id, metric_id, dt) DO UPDATE
SET value = EXCLUDED.value;

WITH station_list AS (
    SELECT id AS station_id, row_number() OVER (ORDER BY id) - 1 AS station_idx
    FROM stations
    WHERE id BETWEEN 101 AND 104
),
field_list AS (
    SELECT id AS field_id, name, row_number() OVER (ORDER BY id) - 1 AS metric_idx
    FROM forecast_fields
    WHERE name IN ('temperature', 'wind_speed', 'humidity', 'pressure', 'precipitation')
),
days AS (
    SELECT generate_series(31, 90) AS d
),
points AS (
    SELECT
        s.station_id,
        s.station_idx,
        f.field_id,
        f.name,
        f.metric_idx,
        d.d,
        date_trunc('day', now()) - (d.d || ' days')::interval + interval '12 hours' AS ts,
        9 + (s.station_idx * 1.7) + (f.metric_idx * 5.5) + (d.d % 7) AS forecast_value,
        9 + (s.station_idx * 1.7) + (f.metric_idx * 5.5) + (d.d % 7)
          + CASE
              WHEN s.station_id = 101 AND f.name = 'wind_speed' AND d.d BETWEEN 31 AND 45 THEN 9
              WHEN s.station_id = 102 AND f.name = 'pressure' AND d.d BETWEEN 46 AND 60 THEN -12
              WHEN s.station_id = 103 AND f.name = 'temperature' AND d.d BETWEEN 61 AND 75 THEN 7
              WHEN s.station_id = 104 AND f.name = 'precipitation' AND d.d BETWEEN 76 AND 90 THEN 13
              ELSE ((d.d + f.metric_idx + s.station_idx) % 7) - 3
            END AS actual_value
    FROM station_list s
    CROSS JOIN field_list f
    CROSS JOIN days d
)
INSERT INTO forecasts (date, field_id, value, interval, station_id, is_archived)
SELECT ts, field_id, forecast_value, '24h', station_id, true
FROM points;

WITH station_list AS (
    SELECT id AS station_id, row_number() OVER (ORDER BY id) - 1 AS station_idx
    FROM stations
    WHERE id BETWEEN 101 AND 104
),
metric_list AS (
    SELECT m.id AS metric_id, ff.name, row_number() OVER (ORDER BY m.id) - 1 AS metric_idx
    FROM metrics m
    JOIN forecast_fields ff ON ff.id = m.forecast_field_id
    WHERE ff.name IN ('temperature', 'wind_speed', 'humidity', 'pressure', 'precipitation')
),
days AS (
    SELECT generate_series(31, 90) AS d
),
points AS (
    SELECT
        s.station_id,
        s.station_idx,
        m.metric_id,
        m.name,
        m.metric_idx,
        d.d,
        date_trunc('day', now()) - (d.d || ' days')::interval + interval '12 hours' AS ts,
        9 + (s.station_idx * 1.7) + (m.metric_idx * 5.5) + (d.d % 7)
          + CASE
              WHEN s.station_id = 101 AND m.name = 'wind_speed' AND d.d BETWEEN 31 AND 45 THEN 9
              WHEN s.station_id = 102 AND m.name = 'pressure' AND d.d BETWEEN 46 AND 60 THEN -12
              WHEN s.station_id = 103 AND m.name = 'temperature' AND d.d BETWEEN 61 AND 75 THEN 7
              WHEN s.station_id = 104 AND m.name = 'precipitation' AND d.d BETWEEN 76 AND 90 THEN 13
              ELSE ((d.d + m.metric_idx + s.station_idx) % 7) - 3
            END AS actual_value
    FROM station_list s
    CROSS JOIN metric_list m
    CROSS JOIN days d
)
INSERT INTO archive (dt, station_id, metric_id, value)
SELECT ts, station_id, metric_id, actual_value
FROM points
ON CONFLICT (station_id, metric_id, dt) DO UPDATE
SET value = EXCLUDED.value;

INSERT INTO alerts (id, title, source, severity, status, started_at, updated_at)
VALUES
    ('dev-alert-kafka-lag', 'Kafka lag above local demo threshold', 'Telemetry Receiver', 'warning', 'open', now() - interval '45 minutes', now() - interval '5 minutes'),
    ('dev-alert-backfill', 'Backfill job waiting for worker', 'ETL/Data Processing', 'info', 'acknowledged', now() - interval '2 hours', now() - interval '20 minutes')
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    source = EXCLUDED.source,
    severity = EXCLUDED.severity,
    status = EXCLUDED.status,
    started_at = EXCLUDED.started_at,
    updated_at = EXCLUDED.updated_at;

INSERT INTO backfill_jobs (
    id,
    status,
    date_from,
    date_to,
    station_id,
    metric,
    calculation_version,
    progress_pct,
    processed_rows,
    failed_rows,
    created_at,
    started_at,
    finished_at
)
VALUES
    ('dev-backfill-completed', 'completed', now() - interval '7 days', now() - interval '1 day', '101', 'wind_speed', 'dev-calc-v1', 100, 840, 0, now() - interval '1 day', now() - interval '1 day', now() - interval '23 hours'),
    ('dev-backfill-running', 'running', now() - interval '3 days', now(), '102', 'pressure', 'dev-calc-v2', 63, 420, 3, now() - interval '30 minutes', now() - interval '20 minutes', null)
ON CONFLICT (id) DO UPDATE
SET status = EXCLUDED.status,
    date_from = EXCLUDED.date_from,
    date_to = EXCLUDED.date_to,
    station_id = EXCLUDED.station_id,
    metric = EXCLUDED.metric,
    calculation_version = EXCLUDED.calculation_version,
    progress_pct = EXCLUDED.progress_pct,
    processed_rows = EXCLUDED.processed_rows,
    failed_rows = EXCLUDED.failed_rows,
    created_at = EXCLUDED.created_at,
    started_at = EXCLUDED.started_at,
    finished_at = EXCLUDED.finished_at;

COMMIT;
