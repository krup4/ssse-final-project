CREATE TABLE IF NOT EXISTS worst_errors
(
    id String,
    station_id String,
    station_name String,
    region_id String,
    region_name String,
    parameter LowCardinality(String),
    metric LowCardinality(String),
    forecast_value Float64,
    actual_value Float64,
    absolute_error Float64,
    error_pct Float64,
    observed_at DateTime64(3, 'UTC'),
    backfill_version String DEFAULT 'calc-v1'
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(observed_at)
ORDER BY (observed_at, region_id, station_id, metric, absolute_error);

CREATE TABLE IF NOT EXISTS parameter_errors
(
    id String,
    parameter LowCardinality(String),
    station_id String,
    station_name String,
    region_id String,
    region_name String,
    forecast_value Float64,
    actual_value Float64,
    absolute_error Float64,
    error_pct Float64,
    contribution_pct Float64,
    samples UInt64,
    observed_at DateTime64(3, 'UTC'),
    backfill_version String DEFAULT 'calc-v1'
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(observed_at)
ORDER BY (observed_at, region_id, station_id, parameter, absolute_error);

CREATE TABLE IF NOT EXISTS parameter_error_trend
(
    timestamp DateTime64(3, 'UTC'),
    parameter LowCardinality(String),
    region_id String,
    station_id String,
    absolute_error Float64,
    mae Float64,
    samples UInt64,
    backfill_version String DEFAULT 'calc-v1'
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, region_id, station_id, parameter);

CREATE TABLE IF NOT EXISTS station_series
(
    timestamp DateTime64(3, 'UTC'),
    station_id String,
    metric LowCardinality(String),
    forecast Float64,
    actual Float64,
    absolute_error Float64,
    backfill_version String DEFAULT 'calc-v1'
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, station_id, metric);

CREATE TABLE IF NOT EXISTS daily_metrics
(
    date Date,
    region_id String,
    region_name String,
    station_id String,
    station_name String,
    metric LowCardinality(String),
    mae Float64,
    rmse Float64,
    samples UInt64,
    worst_error_today Float64,
    active_stations UInt32,
    degraded_stations UInt32,
    offline_stations UInt32,
    kafka_lag Int64,
    request_rate Float64,
    p95_latency_ms Float64,
    backfill_version String DEFAULT 'calc-v1'
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(date)
ORDER BY (date, region_id, station_id, metric);
