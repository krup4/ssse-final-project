# Local demo data

This folder is local-only test tooling. It is not imported by the application,
not copied into service images, and should not be used for production deploys.

## Start stack

```bash
docker compose up --build -d
```

`docker-compose.yml` starts local infrastructure, runs Core API migrations,
applies `dev/seed/postgres_demo.sql` and `dev/seed/archive_history_demo.sql`,
creates Redpanda topics, and publishes a small telemetry batch to Kafka
automatically. The archive seed creates forecast/fact pairs across 180 days so
history and time-series charts have multiple dates.

Frontend: `http://localhost:3000`

Demo logins:

- `admin` / `password`
- `analyst` / `password`
- `operator` / `password`
- `viewer` / `password`

## Re-seed PostgreSQL manually

The compose stack already does this on startup. Run it manually only when you
want to reset demo rows while containers are already running.

```bash
docker compose exec -T postgres psql -U weather -d weather_accuracy < dev/seed/postgres_demo.sql
```

## Publish more Kafka telemetry

The compose stack already publishes an initial batch. Run this manually when
you want more actual-weather events in `actual-weather.raw.v1`.

```bash
./dev/seed/publish_kafka_actual_weather.sh
```

```bash
BATCHES=10 ./dev/seed/publish_kafka_actual_weather.sh
```

## Clean dev data

```bash
docker compose exec -T postgres psql -U weather -d weather_accuracy < dev/seed/clean_postgres_demo.sql
```
