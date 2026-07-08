-- V1__init.sql
-- Initial database schema for Weather Service

-- Create stations table
CREATE TABLE IF NOT EXISTS stations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    latitude FLOAT NOT NULL CHECK (latitude >= -90 AND latitude <= 90),
    longitude FLOAT NOT NULL CHECK (longitude >= -180 AND longitude <= 180),
    region VARCHAR(255) DEFAULT '',
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indices for stations
CREATE INDEX IF NOT EXISTS idx_stations_name ON stations(name);
CREATE INDEX IF NOT EXISTS idx_stations_active ON stations(active);

-- Create weather_measurements table
CREATE TABLE IF NOT EXISTS weather_measurements (
    id SERIAL PRIMARY KEY,
    station_id INTEGER NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    temperature FLOAT CHECK (temperature >= -100 AND temperature <= 100),
    humidity FLOAT CHECK (humidity >= 0 AND humidity <= 100),
    pressure FLOAT CHECK (pressure >= 300 AND pressure <= 1100),
    wind_speed FLOAT CHECK (wind_speed >= 0 AND wind_speed <= 200),
    wind_direction FLOAT CHECK (wind_direction >= 0 AND wind_direction <= 360),
    precipitation FLOAT CHECK (precipitation >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indices for weather_measurements
CREATE INDEX IF NOT EXISTS idx_weather_station_id ON weather_measurements(station_id);
CREATE INDEX IF NOT EXISTS idx_weather_timestamp ON weather_measurements(timestamp);
CREATE INDEX IF NOT EXISTS idx_weather_created_at ON weather_measurements(created_at);
CREATE INDEX IF NOT EXISTS idx_weather_station_timestamp ON weather_measurements(station_id, timestamp);
