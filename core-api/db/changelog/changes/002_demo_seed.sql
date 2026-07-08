--liquibase formatted sql

--changeset core-api:002-demo-seed context:demo
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
ON CONFLICT (login) DO NOTHING;

INSERT INTO forecast_fields (name)
VALUES ('temperature'), ('wind_speed'), ('humidity'), ('pressure'), ('precipitation'), ('wind_gust')
ON CONFLICT (name) DO NOTHING;

INSERT INTO metrics (forecast_field_id, name)
SELECT id, name
FROM forecast_fields
ON CONFLICT (forecast_field_id, name) DO NOTHING;

INSERT INTO stations (id, name, lon, lat, is_active)
VALUES
    (1, 'Ryazan Field', 39.735, 54.626, true),
    (2, 'Caspian Steppe', 48.041, 46.349, true),
    (3, 'Murmansk Port', 33.082, 68.958, false)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('stations', 'id'), (SELECT max(id) FROM stations));

WITH station_list AS (
    SELECT id AS station_id
    FROM stations
),
field_list AS (
    SELECT id AS field_id, name, row_number() OVER (ORDER BY id) - 1 AS metric_idx
    FROM forecast_fields
    WHERE name IN ('temperature', 'wind_speed', 'humidity', 'pressure', 'precipitation')
),
hours AS (
    SELECT generate_series(0, 35) AS h
),
points AS (
    SELECT
        s.station_id,
        f.field_id,
        f.name,
        f.metric_idx,
        h.h,
        date_trunc('hour', now()) - (h.h || ' hours')::interval AS ts,
        10 + (f.metric_idx * 7) + (h.h % 5) AS forecast_value,
        10 + (f.metric_idx * 7) + (h.h % 5) + (((h.h + f.metric_idx) % 9) - 4) AS actual_value
    FROM station_list s
    CROSS JOIN field_list f
    CROSS JOIN hours h
)
INSERT INTO forecasts (date, field_id, value, interval, station_id, is_archived)
SELECT ts, field_id, forecast_value, '6h', station_id, h > 5
FROM points;

WITH station_list AS (
    SELECT id AS station_id
    FROM stations
),
metric_list AS (
    SELECT m.id AS metric_id, ff.name, row_number() OVER (ORDER BY m.id) - 1 AS metric_idx
    FROM metrics m
    JOIN forecast_fields ff ON ff.id = m.forecast_field_id
    WHERE ff.name IN ('temperature', 'wind_speed', 'humidity', 'pressure', 'precipitation')
),
hours AS (
    SELECT generate_series(0, 35) AS h
),
points AS (
    SELECT
        s.station_id,
        m.metric_id,
        m.metric_idx,
        h.h,
        date_trunc('hour', now()) - (h.h || ' hours')::interval AS ts,
        10 + (m.metric_idx * 7) + (h.h % 5) + (((h.h + m.metric_idx) % 9) - 4) AS actual_value
    FROM station_list s
    CROSS JOIN metric_list m
    CROSS JOIN hours h
)
INSERT INTO archive (dt, station_id, metric_id, value)
SELECT ts, station_id, metric_id, actual_value
FROM points;

INSERT INTO alerts (id, title, source, severity, status, started_at, updated_at)
VALUES ('a-1', 'Kafka lag above SLO', 'Telemetry Receiver', 'warning', 'open', now() - interval '40 minutes', now() - interval '5 minutes')
ON CONFLICT DO NOTHING;

--rollback DELETE FROM alerts WHERE id = 'a-1';
--rollback DELETE FROM archive;
--rollback DELETE FROM forecasts;
--rollback DELETE FROM stations WHERE id IN (1, 2, 3);
--rollback DELETE FROM metrics;
--rollback DELETE FROM forecast_fields;
--rollback DELETE FROM users WHERE login IN ('admin', 'analyst', 'operator', 'viewer');
--rollback DELETE FROM roles WHERE name IN ('admin', 'analyst', 'operator', 'viewer');
