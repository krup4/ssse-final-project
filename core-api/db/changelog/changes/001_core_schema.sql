--liquibase formatted sql

--changeset core-api:001-core-schema
CREATE TABLE roles (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE users (
    id serial PRIMARY KEY,
    login text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role_id integer NOT NULL REFERENCES roles(id),
    is_active boolean NOT NULL DEFAULT true,
    name text NOT NULL,
    email text NOT NULL,
    last_seen timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_role_id ON users (role_id);
CREATE INDEX idx_users_is_active ON users (is_active);

CREATE TABLE forecast_fields (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE stations (
    id serial PRIMARY KEY,
    name text NOT NULL,
    lon double precision NOT NULL,
    lat double precision NOT NULL,
    is_active boolean NOT NULL DEFAULT true
);
CREATE INDEX idx_stations_is_active ON stations (is_active);

CREATE TABLE forecasts (
    id serial PRIMARY KEY,
    date timestamptz NOT NULL,
    field_id integer NOT NULL REFERENCES forecast_fields(id),
    value double precision NOT NULL,
    interval text NOT NULL,
    station_id integer NOT NULL REFERENCES stations(id),
    is_archived boolean NOT NULL DEFAULT false
);
CREATE INDEX idx_forecasts_match ON forecasts (station_id, field_id, date);
CREATE INDEX idx_forecasts_is_archived ON forecasts (is_archived);

CREATE TABLE metrics (
    id serial PRIMARY KEY,
    forecast_field_id integer NOT NULL REFERENCES forecast_fields(id),
    name text NOT NULL
);
CREATE INDEX idx_metrics_forecast_field_id ON metrics (forecast_field_id);
CREATE UNIQUE INDEX idx_metrics_field_name ON metrics (forecast_field_id, name);

CREATE TABLE archive (
    id serial PRIMARY KEY,
    dt timestamptz NOT NULL,
    station_id integer NOT NULL REFERENCES stations(id),
    metric_id integer NOT NULL REFERENCES metrics(id),
    value double precision NOT NULL
);
CREATE INDEX idx_archive_match ON archive (station_id, metric_id, dt);
CREATE UNIQUE INDEX idx_archive_unique_point ON archive (station_id, metric_id, dt);

CREATE TABLE alerts (
    id varchar(80) PRIMARY KEY,
    title text NOT NULL,
    source text NOT NULL,
    severity varchar(32) NOT NULL,
    status varchar(32) NOT NULL,
    started_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);
CREATE INDEX idx_alerts_source ON alerts (source);
CREATE INDEX idx_alerts_severity ON alerts (severity);
CREATE INDEX idx_alerts_status ON alerts (status);

CREATE TABLE backfill_jobs (
    id varchar(80) PRIMARY KEY,
    status varchar(32) NOT NULL,
    date_from timestamptz NOT NULL,
    date_to timestamptz NOT NULL,
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
CREATE INDEX idx_backfill_jobs_status ON backfill_jobs (status);
CREATE INDEX idx_backfill_jobs_date_from ON backfill_jobs (date_from);
CREATE INDEX idx_backfill_jobs_date_to ON backfill_jobs (date_to);
CREATE INDEX idx_backfill_jobs_station_id ON backfill_jobs (station_id);
CREATE INDEX idx_backfill_jobs_metric ON backfill_jobs (metric);
CREATE INDEX idx_backfill_jobs_calculation_version ON backfill_jobs (calculation_version);

--rollback DROP TABLE IF EXISTS backfill_jobs;
--rollback DROP TABLE IF EXISTS alerts;
--rollback DROP TABLE IF EXISTS archive;
--rollback DROP TABLE IF EXISTS metrics;
--rollback DROP TABLE IF EXISTS forecasts;
--rollback DROP TABLE IF EXISTS stations;
--rollback DROP TABLE IF EXISTS forecast_fields;
--rollback DROP TABLE IF EXISTS users;
--rollback DROP TABLE IF EXISTS roles;
