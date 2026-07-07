# Weather Forecast Accuracy Core API

Core API для учебного проекта сервиса поиска точек, где ошибка прогноза погоды максимальна.

## Структура

- `core-api/` - реализованный Core API на Go, Gin, GORM.
- `openapi.yml` - контракт API для UI аналитики.

Внутри `core-api`:

- `cmd/core-api` - точка входа приложения.
- `internal/adapters/httpgin` - HTTP layer: router, middleware, handlers по группам OpenAPI.
- `internal/service` - сервисный слой/use cases: auth, analytics, alerts, users, dictionaries, backfills.
- `internal/domain` - доменные модели, ошибки и repository interfaces.
- `internal/adapters/postgres` - GORM-модели и PostgreSQL repositories, разнесённые по сущностям.
- `internal/adapters/clickhouse` - аналитический adapter для Gold-витрин ClickHouse.
- `internal/adapters/kafka` - Kafka consumer и decoder фактической погоды.
- `internal/adapters/redisstore` - Redis-backed rate limiting, response cache, Kafka event deduplication.
- `db/changelog` - Liquibase changelog для PostgreSQL-схемы и demo seed.
- `internal/config` - env-конфигурация.
- `internal/platform` - технические утилиты: logger, metrics registry.

## Core API

```bash
cd core-api
go mod download
go run ./cmd/core-api
```

По умолчанию API слушает `:8080`, база ожидается по:

```text
postgres://weather:weather@localhost:5432/weather_accuracy?sslmode=disable
```

Основные переменные окружения:

- `DATABASE_URL`
- `ANALYTICS_BACKEND`
- `CLICKHOUSE_DSN`
- `CLICKHOUSE_CONNECT_TIMEOUT`
- `CLICKHOUSE_CONNECT_RETRIES`
- `CLICKHOUSE_CONNECT_BACKOFF`
- `CLICKHOUSE_MAX_OPEN_CONNS`
- `CLICKHOUSE_MAX_IDLE_CONNS`
- `JWT_SECRET`
- `HTTP_REQUEST_TIMEOUT`
- `DB_CONNECT_TIMEOUT`
- `DB_CONNECT_RETRIES`
- `DB_CONNECT_BACKOFF`
- `DB_MAX_OPEN_CONNS`
- `DB_MAX_IDLE_CONNS`
- `DB_CONN_MAX_LIFETIME`
- `DB_CONN_MAX_IDLE_TIME`
- `DB_PING_TIMEOUT`
- `KAFKA_BROKERS`
- `KAFKA_ACTUAL_WEATHER_TOPIC`
- `KAFKA_BACKFILL_JOBS_TOPIC`
- `KAFKA_ENABLED`
- `REDIS_ENABLED`
- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `REDIS_KEY_PREFIX`
- `REDIS_CACHE_TTL`
- `REDIS_DEDUP_TTL`
- `REDIS_RATE_LIMIT_TTL`
- `AUTO_MIGRATE`

`AUTO_MIGRATE` по умолчанию выключен. Нормальный путь для схемы БД - Liquibase-миграции до запуска приложения.

## Миграции БД

Liquibase changelog:

```text
core-api/db/changelog/db.changelog-master.yaml
```

Пример запуска из `core-api/`:

```bash
liquibase \
  --changeLogFile=db/changelog/db.changelog-master.yaml \
  --url=jdbc:postgresql://localhost:5432/weather_accuracy \
  --username=weather \
  --password=weather \
  update
```

Для локальных demo-данных:

```bash
liquibase \
  --changeLogFile=db/changelog/db.changelog-master.yaml \
  --url=jdbc:postgresql://localhost:5432/weather_accuracy \
  --username=weather \
  --password=weather \
  --contexts=demo \
  update
```

Пример properties-файла:

```text
core-api/db/liquibase.properties.example
```

## Аналитика в ClickHouse

По умолчанию `ANALYTICS_BACKEND=postgres`, чтобы сервис можно было запустить локально без ClickHouse.
Для режима по ТЗ:

```bash
ANALYTICS_BACKEND=clickhouse
CLICKHOUSE_DSN=clickhouse://weather:weather@clickhouse:9000/weather_dwh
```

Core API читает Gold-витрины:

- `worst_errors`
- `parameter_errors`
- `parameter_error_trend`
- `station_series`
- `daily_metrics`

Ожидаемый DDL для витрин лежит здесь:

```text
core-api/db/clickhouse/001_gold_views.sql
```

## Устойчивость к нестабильной сети

- PostgreSQL подключается с retry и exponential backoff.
- PostgreSQL connection pool настраивается через env, чтобы контролировать нагрузку на managed DB.
- `/readyz` проверяет PostgreSQL через `PingContext`; Kubernetes сможет убрать pod из балансировки, если БД недоступна.
- HTTP-запросы получают общий timeout через `HTTP_REQUEST_TIMEOUT`.
- Kafka consumer сохраняет сообщение в БД до commit offset и использует retry/backoff при временных ошибках записи.
- `POST /backfills` создаёт job в PostgreSQL и публикует `backfill.requested` event в Kafka topic `backfill.jobs.v1` для ETL/Data Processing service.
- `/metrics` отдаёт Prometheus text format: HTTP request count, 5xx error count, latency histogram, in-flight requests, Kafka processed/error/commit counters.
- Redis используется для distributed rate limiting между pod replicas, короткого cache для частых GET-ручек и dedup actual-weather событий из Kafka.

Демо-логины:

- `admin@weather.local` / `password`
- `analyst@weather.local` / `password`
- `operator@weather.local` / `password`

Пример:

```bash
TOKEN=$(curl -s http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"analyst@weather.local","password":"password"}' | jq -r .token)

curl "http://localhost:8080/api/v1/analytics/worst-errors?dateFrom=2026-07-01&dateTo=2026-07-08&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

## Форматы данных

Forecast поток сейчас считается записанным в PostgreSQL в таблицу `forecast_reading_models`.
Actual weather приходит из Kafka topic `actual-weather.raw.v1`; Core API consumer декодирует JSON и сохраняет нормализованные значения в PostgreSQL.
Backfill jobs публикуются в Kafka topic `backfill.jobs.v1`, чтобы ETL/Data Processing service запускал перерасчёт Gold-витрин.

Текущий JSON actual weather:

```json
{
  "id": "sensor-event-1",
  "stationId": "st-004",
  "observedAt": "2026-07-08T06:00:00Z",
  "temperatureMax": 29.3,
  "windSpeed": 12.1,
  "humidity": 71,
  "pressure": 1009,
  "precipitationTotal": 0.2,
  "source": "sensor-gateway",
  "traceId": "trace-1"
}
```

Когда реальный формат Kafka или схемы forecast изменится, менять нужно адаптеры в `internal/adapters/kafka` и `internal/adapters/postgres`, а не handlers.
