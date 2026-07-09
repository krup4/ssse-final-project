# Weather Forecast Accuracy Analytics

Система аналитики точности прогноза погоды. Проект собирает прогнозы и фактическую погоду по метеостанциям, считает ошибки прогноза по параметрам и метрикам, показывает overview, history, worst errors, parameter errors, графики и карту станций.

## Что делает проект

- Загружает список станций и прогнозы погоды из Yandex Weather GraphQL API.
- Получает фактическую погоду из Yandex Weather API по расписанию.
- Передает фактическую погоду через Kafka topic `actual-weather.raw.v1`.
- Сохраняет прогнозы и рассчитанные ошибки в PostgreSQL.
- Отдает REST API для фронтенда: авторизация, аналитика, станции, пользователи, alerts, backfill jobs.
- Показывает UI с дашбордом, графиками, историей, таблицами ошибок и картой станций.
- Поддерживает локальный запуск через Docker Compose и production-style деплой в Yandex Cloud Kubernetes.

## Архитектура

```text
                  +-------------------------+
                  |        Frontend         |
                  | React + Vite + Nginx    |
                  +-----------+-------------+
                              |
                              | /api/v1
                              v
                  +-------------------------+
                  |        Core API         |
                  | Go + Gin + GORM         |
                  +-----+-------------+-----+
                        |             |
                        |             | consumes
                        v             v
              +----------------+   +--------------------------+
              |   PostgreSQL   |   | Kafka actual-weather...  |
              | domain storage |   +-------------+------------+
              +-------+--------+                 ^
                      ^                          |
                      | writes forecasts         | publishes facts
                      |                          |
        +-------------+------------+   +---------+----------------+
        |    Forecast Service      |   | Actual Weather Feather  |
        | Python + APScheduler     |   | Python + FastAPI        |
        +--------------------------+   +--------------------------+
                      |                          |
                      v                          v
        Yandex Weather GraphQL API       Yandex Weather API
```

### Основные сервисы

`frontend`

React-приложение для операторов и аналитиков. Содержит dashboard, charts, history, errors, parameter errors, stations/users/alerts pages и карту станций на Yandex Maps. Собирается Vite, в production отдается Nginx. Runtime-конфиг для карты генерируется в `/config.js` из переменной `YANDEX_MAPS_API_KEY`, поэтому один и тот же image можно использовать в разных окружениях.

`core-api`

Главный backend на Go. Реализует HTTP API, JWT-авторизацию, RBAC, analytics use cases, CRUD справочников, alerts, backfill jobs, Kafka consumer фактической погоды, Redis cache/rate limit/dedup и Prometheus metrics. Работает с PostgreSQL как основной БД. Может читать аналитические витрины из ClickHouse при `ANALYTICS_BACKEND=clickhouse`, но локально и в текущей рабочей схеме может использовать `ANALYTICS_BACKEND=postgres`.

`forecast-service`

Python-сервис, который инициализирует станции из `forecast-service/db/stations.json`, ходит в Yandex Weather GraphQL API, получает прогнозы и пишет их в таблицу `forecasts`. Запускает обновление сразу при старте и дальше по расписанию через APScheduler в timezone `Europe/Moscow`.

`actual-weather-feather`

Python-сервис для сбора фактической погоды. Читает активные станции из PostgreSQL, запрашивает Yandex Weather API, дедуплицирует события через Redis и публикует измерения в Kafka topic `actual-weather.raw.v1`. Сам не пишет фактические измерения в PostgreSQL: сохранение и расчет ошибок выполняет `core-api` consumer.

`PostgreSQL`

Основное хранилище доменных данных: users, roles, stations, forecast_fields, forecasts, metrics, archive, alerts, backfill_jobs. Таблица `archive` хранит рассчитанные значения ошибок/метрик по фактической погоде и прогнозам.

`Kafka`

Шина событий для фактической погоды и backfill jobs. Локально используется Redpanda, в Yandex Cloud ожидается Managed Kafka.

`Redis / Valkey`

Используется для distributed rate limiting, short-lived API cache, dedup Kafka events и leader lock в `actual-weather-feather`, чтобы несколько replica не запускали одинаковый scheduled collection одновременно.

`ClickHouse`

Опциональное аналитическое хранилище для Gold-витрин (`worst_errors`, `parameter_errors`, `parameter_error_trend`, `station_series`, `daily_metrics`). DDL лежит в `core-api/db/clickhouse/001_gold_views.sql`.

## Поток данных

1. `forecast-service` читает станции из PostgreSQL и пишет прогнозы в `forecasts`.
2. `actual-weather-feather` получает фактическую погоду по станциям и публикует события в Kafka.
3. `core-api` читает Kafka topic `actual-weather.raw.v1`.
4. `core-api` сопоставляет фактические значения с прогнозами по станции, параметру и времени.
5. Рассчитанные ошибки сохраняются в PostgreSQL `archive`.
6. Analytics endpoints читают агрегаты из PostgreSQL или ClickHouse, в зависимости от `ANALYTICS_BACKEND`.
7. `frontend` получает данные через `/api/v1` и отображает dashboard, trends, tables и map.

## Технологии

Backend:

- Go
- Gin
- GORM
- PostgreSQL
- ClickHouse driver
- Kafka client
- Redis client
- JWT auth
- Prometheus-compatible metrics
- Liquibase migrations

Python services:

- Python 3.12
- FastAPI
- asyncpg
- APScheduler
- httpx
- aiokafka
- Redis
- structlog
- pytest

Frontend:

- React 18
- TypeScript
- Vite
- TanStack Query
- React Router
- Recharts
- Lucide React
- Axios
- MSW for mocks
- Yandex Maps JavaScript API
- Nginx runtime serving

Infrastructure:

- Docker / Docker Compose
- Kubernetes manifests
- Yandex Cloud Managed Kubernetes
- Yandex Container Registry
- Managed PostgreSQL
- Managed Kafka
- Managed ClickHouse
- Managed Valkey
- Terraform scaffold
- MinIO in local compose

## Repository Structure

```text
.
├── actual-weather-feather/   # фактическая погода: Yandex Weather -> Kafka
├── core-api/                 # Go Core API, domain, repositories, migrations
├── dev/                      # local-only demo seed and helper scripts
├── forecast-service/         # прогнозы: Yandex Weather GraphQL -> PostgreSQL
├── frontend/                 # React UI, Nginx image, runtime config
├── infra/                    # Terraform scaffold for Yandex Cloud
├── k8s/base/                 # Kubernetes manifests
├── docker-compose.yml        # локальный full stack
└── DEPLOY_YC.md              # краткий деплой в Yandex Cloud
```

## Local Start

Самый простой способ поднять весь локальный стенд:

```bash
docker compose up --build -d
```

Compose поднимает PostgreSQL, Redis, Redpanda, ClickHouse, MinIO, Core API и frontend. Также применяются demo seed scripts из `dev/seed`, создаются Kafka topics и публикуется небольшая пачка фактической погоды.

Frontend будет доступен на:

```text
http://localhost:3000
```

Core API:

```text
http://localhost:8080
```

Demo users:

- `admin` / `password`
- `analyst` / `password`
- `operator` / `password`
- `viewer` / `password`

Для карты станций задайте ключ Яндекс.Карт:

```bash
YANDEX_MAPS_API_KEY=your-yandex-maps-api-key docker compose up --build -d frontend
```

## Local Development

### Core API

```bash
cd core-api
go mod download
go run ./cmd/core-api
```

Полезные переменные:

- `DATABASE_URL`
- `ANALYTICS_BACKEND`
- `JWT_SECRET`
- `KAFKA_ENABLED`
- `KAFKA_BROKERS`
- `REDIS_ENABLED`
- `REDIS_ADDR`
- `AUTO_MIGRATE`
- `CLICKHOUSE_DSN`

Тесты:

```bash
cd core-api
go test ./...
```

### Frontend

```bash
cd frontend
npm install
cat > public/config.js <<'EOF'
window.__APP_CONFIG__ = {
  yandexMapsApiKey: "your-yandex-maps-api-key"
};
EOF
npm run dev
```

Проверка production build:

```bash
cd frontend
npm run build
```

### Actual Weather Feather

```bash
cd actual-weather-feather
python -m pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8080
```

Основные переменные:

- `DATABASE_URL` или `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `KAFKA_BROKERS`
- `KAFKA_TOPIC`
- `REDIS_URL` / Redis connection variables
- `YANDEX_WEATHER_API_KEY`
- `YANDEX_WEATHER_URL`
- `UPDATE_INTERVAL_SECONDS`
- `REQUEST_TIMEOUT`

Тесты:

```bash
cd actual-weather-feather
python -m pytest
```

### Forecast Service

`forecast-service` запускается как scheduler process. Он использует:

- `DATABASE_URL`
- `FORECAST_API_KEY`
- `URL`
- `FORECAST_RATE_LIMIT_RPS`

При старте сервис заполняет базовые станции и сразу выполняет обновление прогнозов.

## Database Migrations

PostgreSQL schema описана Liquibase changelog:

```text
core-api/db/changelog/db.changelog-master.yaml
```

Пример ручного запуска:

```bash
cd core-api
liquibase \
  --changeLogFile=db/changelog/db.changelog-master.yaml \
  --url=jdbc:postgresql://localhost:5432/weather_accuracy \
  --username=weather \
  --password=weather \
  update
```

В локальном Docker Compose включен `AUTO_MIGRATE=true` для `core-api`, поэтому схема создается автоматически.

## API and Auth

`core-api` публикует REST API под `/api/v1`.

Пример логина:

```bash
TOKEN=$(curl -s http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"login":"analyst","password":"password"}' | jq -r .token)
```

Пример запроса аналитики:

```bash
curl "http://localhost:8080/api/v1/analytics/worst-errors?dateFrom=2026-07-01&dateTo=2026-07-08&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

Health endpoints:

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`

## Kubernetes and Yandex Cloud

Production-oriented scaffold находится в:

- `infra/` - Terraform for Yandex Cloud resources
- `k8s/base/` - Kubernetes manifests

Ожидаемая production-инфраструктура:

- Managed Kubernetes
- Container Registry
- Managed PostgreSQL
- Managed Kafka
- Managed ClickHouse
- Managed Valkey

Минимальный порядок деплоя:

1. Заполнить `infra/envs/prod/terraform.tfvars`.
2. Выполнить `terraform init`, `terraform plan`, `terraform apply` в `infra/envs/prod`.
3. Собрать и запушить images в Container Registry.
4. Создать Kubernetes secrets для сервисов.
5. Применить manifests:

```bash
kubectl apply -k k8s/base
```

Runtime key для карты:

```bash
kubectl create secret generic frontend-secret \
  -n weather \
  --from-literal=YANDEX_MAPS_API_KEY='your-key' \
  --dry-run=client -o yaml | kubectl apply -f -

kubectl rollout restart -n weather deploy/frontend
```

## Scaling and Reliability

- `core-api` рассчитан на несколько replicas, имеет readiness/liveness/startup probes, HPA и PDB.
- `actual-weather-feather` может работать в нескольких replicas; Redis leader lock предотвращает duplicate scheduler cycles.
- PostgreSQL connection pool, retries и backoff настраиваются через env.
- Redis используется для distributed rate limiting и dedup.
- Kafka consumer сохраняет данные до commit offset, чтобы не терять события при временных ошибках.
- `/readyz` проверяет доступность зависимостей, чтобы Kubernetes исключал неготовые pod'ы из балансировки.
- Logs пишутся в stdout, метрики доступны в Prometheus text format.

## Notes

- `dev/` содержит только локальные seed scripts и не предназначен для production.
- `ANALYTICS_BACKEND=postgres` удобен для локального запуска и текущего простого production path.
- `ANALYTICS_BACKEND=clickhouse` включает чтение Gold-витрин ClickHouse, если они корректно мигрированы и синхронизированы.
- Yandex Maps key не вшивается в frontend image: контейнер генерирует `/config.js` на старте.
