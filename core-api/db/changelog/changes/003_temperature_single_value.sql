--liquibase formatted sql

--changeset weather-accuracy:003-temperature-single-value
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'actual_weather_reading_models'
          AND column_name = 'temperature_max'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'actual_weather_reading_models'
          AND column_name = 'temperature'
    ) THEN
        ALTER TABLE actual_weather_reading_models RENAME COLUMN temperature_max TO temperature;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'actual_weather_reading_models'
          AND column_name = 'temperature_min'
    ) THEN
        ALTER TABLE actual_weather_reading_models DROP COLUMN temperature_min;
    END IF;
END $$;
--rollback ALTER TABLE actual_weather_reading_models RENAME COLUMN temperature TO temperature_max;
--rollback ALTER TABLE actual_weather_reading_models ADD COLUMN temperature_min double precision;
