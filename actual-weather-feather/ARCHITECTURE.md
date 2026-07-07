# Архитектура Actual Weather Service (Python)

Этот документ описывает новый стек и подсистемы, которые должен реализовать следующий агент. Сервис переписывается на **FastAPI** + **asyncio**, сохраняя прежнюю ответственность:
1. Получать активные станции из PostgreSQL.
2. Каждые полный час запрашивать актуальную погоду из Open-Meteo.
3. Валидировать и нормализовывать данные.
4. Сохранять измерения в PostgreSQL.
5. Публиковать события в Kafka (`weather.actual`).
6. Дедуплицировать через Redis.
7. Экспонировать health/metrics + CRUD для станций.

Все сущности должны реализовываться через `asyncio`, зависимостями нужно управлять через DI (конструкторы, фабрики, contextvars). Экземпляры клиентов и бд-сессий должны быть единожды сконфигурированы и распределены в `app/core`.

---

## Технологический стек

| Компонент | Стек |
|-----------|------|
| HTTP API | FastAPI + uvicorn (httpx для внешних запросов) |
| Конфигурация | Pydantic BaseSettings |
| PostgreSQL | SQLAlchemy async + asyncpg или прямой asyncpg + SQLAlchemy models |
| Миграции | Alembic (SQL-скрипты можно хранить в `migrations/`) |
| Kafka | aiokafka Producer |
| Redis | aioredis (для дедупликации сообщений по ключу `station_id:timestamp`) |
| HTTP-клиент | httpx.AsyncClient с retry и exponential backoff |
| Scheduler | `asyncio` задача + `asyncio.sleep` до следующего полного часа |
| Metrics | prometheus_client (FastAPI instrumentation) |
| Logging | structlog или logging + contextual filters + JSON output |

---

## Компонентная диаграмма

```text ARCHITECTURE.md
                  +---------------------+                +---------+
            +---->|      Scheduler     |--------------->| Open-    |
            |     | (ждёт полного часа)|                | Meteo API|
            |     +----+----------------+                +---------+
            |          |                                     |
            |          V                                     V
            |     +----------+  +-----------+          +------------+
            |     | Services |->| Repositories|<--------| PostgreSQL |
            |     +----------+  +-----------+          +------------+
            |          |                                     |
            |          V                                     |
 +------------+---+      |                               +----v-----+
 | FastAPI / API   |<-----+                               |Kafka      |
 | health/metrics  |                                      |Producer   |
 +-----------------+                                      +----------+
```

---

## Слой api

- `app/api/routes/health.py` — `GET /health`, проверяет PostgreSQL, Redis, Kafka connectivity (можно mock-ответы пока логика не написана).
- `app/api/routes/metrics.py` — `GET /metrics`, использует `prometheus_client.generate_latest()`.
- `app/api/routes/stations.py` — CRUD (GET список, POST, PUT, DELETE). Каждый маршрут ожидает `Session` и `StationService`/`StationRepository`.
- `app/api/deps.py` — зависимости (settings, repositories, kafka/redis_clients).

***Инструкция для следующего агента***: Добавьте `APIRouter`, но пока оставьте тело ручек `NotImplementedError`; опишите контракт входных/выходных `pydantic` схем.

---

## Слой services

`app/services/weather.py` отвечает за:
- получение списка активных станций;
- запрос к Open-Meteo (через `AsyncClient`);
- валидация температуры, давления, ветра и humidity (границы в `config`);
- нормализация (`timestamp` должен быть округлён до часа);
- дедупликация (см. Redis);
- сохранение через `WeatherRepository`;
- публикация события в Kafka (топик `weather.actual`).

`StationService` концентрируется на CRUD.

Пока можно реализовать сервисы как `class WeatherService:` с методами `async def collect(...) -> None` и `async def publish(...) -> None`, а тела оставить `raise NotImplementedError`. Важно описать интерфейс и зависимости.

---

## Scheduler

`app/scheduler/__init__.py` — `async def run_scheduler(app: FastAPI):`.
Алгоритм:
1. Вычислить `sleep_seconds = seconds_until_next_hour()`.
2. `await asyncio.sleep(sleep_seconds)`.
3. Загрузить активные станции (через `StationRepository`).
4. Для каждой станции вызвать `asyncio.create_task(weather_service.process_station(station))`.
5. После запуска всех задач ждать `asyncio.gather(*tasks)` или использовать semaphore for throughput.
6. Повторить цикл, пока приложения живо.

Scheduler должен быть `startup event` FastAPI.

---

## Клиент Open-Meteo

- размещается в `app/clients/open_meteo.py`;
- строит URL `https://api.open-meteo.com/v1/forecast?latitude={lat}&longitude={lon}&current_weather=true&hourly=...`;
- использует `httpx.AsyncClient` с `timeout`, `retry` (exponential backoff: 1s, 2s, 4s, 8s);
- декодирует JSON в `pydantic` модель `OpenMeteoResponse`.

Пока можно реализовать метод `async def fetch(lat, lon)` и бросать `NotImplementedError`, но описать структуру данных.

---

## Репозитории

- `app/repositories/station.py` — методы `get_active`, `create`, `update`, `delete`.
- `app/repositories/weather.py` — `save_measurement`, `get_latest_for_station`.
- Для будущей реализации использует `async_session` (SQLAlchemy or asyncpg connection pool). Пока описать интерфейсы и запускаемые методы.

---

## Мониторинг и дедупликация

- `prometheus_client` метрики: `weather_requests_total`, `weather_saved_total`, `errors_total`, `stations_processed_total`, `retry_total`, `request_duration_seconds`.
- Redis (aioredis) хранит ключ `weather:{station_id}:{timestamp}` с TTL 65 мин для дедупликации.
- Логи должны быть JSON-структурой с полями `station_id`, `duration_ms`, `status`, `error`.

---

## Kafka

- Topic: `weather.actual`.
- Сообщение:

```json
{
  "station_id": 1,
  "timestamp": "2026-06-01T15:00:00Z",
  "temperature": 18.6,
  "humidity": 71,
  "pressure": 760,
  "wind_speed": 5.2,
  "wind_direction": 180,
  "precipitation": 0.0
}
```

- Продюсер конфигурируется через `app/core/kafka.py`; пока можно описать класс `AsyncKafkaPublisher` и методы `async def publish(message: WeatherMessage)`.

---

## Контейнеризация и локальная инфраструктура

- `Dockerfile` — multi-stage сборка: base `python:3.12-slim`, устанавливается pip-зависимости, копируется `app/`, `scripts/`, `migrations/`, `requirements.txt`.
- `docker-compose.yml` — PostgreSQL, Kafka (можно оставить из Go версии), Redis и новый сервис `actual-weather-python`.
- `scripts/docker-entrypoint.sh` — применяет миграции (`alembic upgrade head` или SQL-скрипты), затем запускает `uvicorn`.

Добавьте placeholder команды и комментарии, чтобы позже заменить на реальные миграции.

---

## Что оставить следующему агенту

1. **Создать структуру проекта** (`app/`, `tests/`, `configs/`).
2. **Описать Pydantic схемы** (`StationCreate`, `StationRead`, `WeatherMeasurement`).
3. **Документировать API** в `README`/`ARCHITECTURE`.
4. **Добавить pytest-skeleton** для каждого слоя (тесты пока `pytest.mark.skip` или `NotImplementedError`).
5. **Оставить TODO** в коде, но не решайте их сейчас. В документации укажите на требуемые шаги.

После реализации всех модулей вернуть сюда обновлённые документы и добавить примеры cURL/HTTP для ручек.
