--liquibase formatted sql

--changeset core-api:001-core-schema
CREATE TABLE user_models (
    id varchar(64) PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL,
    password_hash text NOT NULL,
    role varchar(32) NOT NULL,
    last_seen timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_user_models_email ON user_models (lower(email));
CREATE INDEX idx_user_models_role ON user_models (role);

CREATE TABLE region_models (
    id varchar(64) PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE station_models (
    id varchar(64) PRIMARY KEY,
    name text NOT NULL,
    region_id varchar(64) NOT NULL REFERENCES region_models(id),
    lat double precision NOT NULL,
    lon double precision NOT NULL,
    status varchar(32) NOT NULL,
    active_sensors integer NOT NULL,
    last_telemetry_at timestamptz NOT NULL
);

CREATE INDEX idx_station_models_region_id ON station_models (region_id);
CREATE INDEX idx_station_models_status ON station_models (status);
CREATE INDEX idx_station_models_last_telemetry_at ON station_models (last_telemetry_at);

CREATE TABLE forecast_reading_models (
    id varchar(80) PRIMARY KEY,
    station_id varchar(64) NOT NULL REFERENCES station_models(id),
    metric varchar(64) NOT NULL,
    value double precision NOT NULL,
    forecast_at timestamptz NOT NULL,
    target_at timestamptz NOT NULL,
    source text NOT NULL,
    raw_payload jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_forecast_reading_models_forecast_at ON forecast_reading_models (forecast_at);
CREATE INDEX idx_forecast_reading_models_target_at ON forecast_reading_models (target_at);
CREATE INDEX idx_forecast_match ON forecast_reading_models (station_id, metric, target_at);

CREATE TABLE actual_weather_reading_models (
    id varchar(120) PRIMARY KEY,
    station_id varchar(64) NOT NULL REFERENCES station_models(id),
    observed_at timestamptz NOT NULL,
    temperature_min double precision,
    temperature_max double precision,
    precipitation_total double precision,
    wind_speed double precision,
    wind_gust double precision,
    humidity double precision,
    pressure double precision,
    source text NOT NULL,
    trace_id text,
    raw_payload jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_actual_station_observed ON actual_weather_reading_models (station_id, observed_at);
CREATE INDEX idx_actual_weather_reading_models_trace_id ON actual_weather_reading_models (trace_id);

CREATE TABLE alert_models (
    id varchar(80) PRIMARY KEY,
    title text NOT NULL,
    source text NOT NULL,
    severity varchar(32) NOT NULL,
    status varchar(32) NOT NULL,
    started_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX idx_alert_models_source ON alert_models (source);
CREATE INDEX idx_alert_models_severity ON alert_models (severity);
CREATE INDEX idx_alert_models_status ON alert_models (status);

CREATE TABLE backfill_job_models (
    id varchar(80) PRIMARY KEY,
    status varchar(32) NOT NULL,
    date_from timestamptz NOT NULL,
    date_to timestamptz NOT NULL,
    region_id varchar(64),
    station_id varchar(64),
    metric varchar(64),
    calculation_version text NOT NULL,
    progress_pct double precision NOT NULL DEFAULT 0,
    processed_rows bigint NOT NULL DEFAULT 0,
    failed_rows bigint NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,
    started_at timestamptz,
    finished_at timestamptz
);

CREATE INDEX idx_backfill_job_models_status ON backfill_job_models (status);
CREATE INDEX idx_backfill_job_models_date_from ON backfill_job_models (date_from);
CREATE INDEX idx_backfill_job_models_date_to ON backfill_job_models (date_to);
CREATE INDEX idx_backfill_job_models_region_id ON backfill_job_models (region_id);
CREATE INDEX idx_backfill_job_models_station_id ON backfill_job_models (station_id);
CREATE INDEX idx_backfill_job_models_metric ON backfill_job_models (metric);
CREATE INDEX idx_backfill_job_models_calculation_version ON backfill_job_models (calculation_version);

--rollback DROP TABLE IF EXISTS backfill_job_models;
--rollback DROP TABLE IF EXISTS alert_models;
--rollback DROP TABLE IF EXISTS actual_weather_reading_models;
--rollback DROP TABLE IF EXISTS forecast_reading_models;
--rollback DROP TABLE IF EXISTS station_models;
--rollback DROP TABLE IF EXISTS region_models;
--rollback DROP TABLE IF EXISTS user_models;
