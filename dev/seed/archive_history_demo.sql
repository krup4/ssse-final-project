\set ON_ERROR_STOP on

BEGIN;

DELETE FROM archive
WHERE station_id BETWEEN 101 AND 104
  AND dt < date_trunc('day', now()) - interval '90 days';

DELETE FROM forecasts
WHERE station_id BETWEEN 101 AND 104
  AND date < date_trunc('day', now()) - interval '90 days';

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
    SELECT generate_series(91, 180) AS d
),
hours AS (
    SELECT unnest(array[0, 6, 12, 18]) AS hour_offset
),
points AS (
    SELECT
        s.station_id,
        s.station_idx,
        f.field_id,
        f.name,
        f.metric_idx,
        d.d,
        h.hour_offset,
        date_trunc('day', now()) - (d.d || ' days')::interval + (h.hour_offset || ' hours')::interval AS ts,
        8 + (s.station_idx * 2.1) + (f.metric_idx * 4.8) + (d.d % 11) + (h.hour_offset / 6) AS forecast_value,
        8 + (s.station_idx * 2.1) + (f.metric_idx * 4.8) + (d.d % 11) + (h.hour_offset / 6)
          + CASE
              WHEN s.station_id = 101 AND f.name = 'wind_speed' AND d.d BETWEEN 91 AND 115 THEN 10
              WHEN s.station_id = 102 AND f.name = 'pressure' AND d.d BETWEEN 116 AND 135 THEN -15
              WHEN s.station_id = 103 AND f.name = 'temperature' AND d.d BETWEEN 136 AND 160 THEN 8
              WHEN s.station_id = 104 AND f.name = 'precipitation' AND d.d BETWEEN 161 AND 180 THEN 14
              ELSE ((d.d + h.hour_offset + f.metric_idx + s.station_idx) % 9) - 4
            END AS actual_value
    FROM station_list s
    CROSS JOIN field_list f
    CROSS JOIN days d
    CROSS JOIN hours h
)
INSERT INTO forecasts (date, field_id, value, interval, station_id, is_archived)
SELECT ts, field_id, forecast_value, '6h', station_id, true
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
    SELECT generate_series(91, 180) AS d
),
hours AS (
    SELECT unnest(array[0, 6, 12, 18]) AS hour_offset
),
points AS (
    SELECT
        s.station_id,
        s.station_idx,
        m.metric_id,
        m.name,
        m.metric_idx,
        d.d,
        h.hour_offset,
        date_trunc('day', now()) - (d.d || ' days')::interval + (h.hour_offset || ' hours')::interval AS ts,
        8 + (s.station_idx * 2.1) + (m.metric_idx * 4.8) + (d.d % 11) + (h.hour_offset / 6)
          + CASE
              WHEN s.station_id = 101 AND m.name = 'wind_speed' AND d.d BETWEEN 91 AND 115 THEN 10
              WHEN s.station_id = 102 AND m.name = 'pressure' AND d.d BETWEEN 116 AND 135 THEN -15
              WHEN s.station_id = 103 AND m.name = 'temperature' AND d.d BETWEEN 136 AND 160 THEN 8
              WHEN s.station_id = 104 AND m.name = 'precipitation' AND d.d BETWEEN 161 AND 180 THEN 14
              ELSE ((d.d + h.hour_offset + m.metric_idx + s.station_idx) % 9) - 4
            END AS actual_value
    FROM station_list s
    CROSS JOIN metric_list m
    CROSS JOIN days d
    CROSS JOIN hours h
)
INSERT INTO archive (dt, station_id, metric_id, value)
SELECT ts, station_id, metric_id, actual_value
FROM points
ON CONFLICT (station_id, metric_id, dt) DO UPDATE
SET value = EXCLUDED.value;

COMMIT;
