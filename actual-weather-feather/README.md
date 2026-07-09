# Actual Weather Service (Python)

Микросервис получает фактические метеоданные из API Яндекс.Погоды и публикует их в Kafka (`actual-weather.raw.v1`). Он не сохраняет измерения в PostgreSQL (в базе читается только таблица `stations`), не выполняет сравнение с прогнозами и не взаимодействует с аналитикой напрямую — всё происходит через Kafka.

## Что делает сервис

1. Загружает `stations` (поля `id`, `name`, `lat`, `lon`, опционально фильтруя по `name`) из PostgreSQL.
2. По списку станций обращается к Яндекс.Погоде, получает `fact` и маппит его в `WeatherMeasurement`.
3. Дедуплицирует данные по Redis (`station:{id}:last_timestamp`).
4. Публикует событие-контейнер `WeatherKafkaMessage` в топик `actual-weather.raw.v1`.
5. Зацикливает сбор на интервале (`UPDATE_INTERVAL_SECONDS`) и логгирует ключевые шаги (запуск цикла, количество станций, запросы к API, публикации в Kafka, ошибки).

## Архитектура компонентов

- `Scheduler` (`app/scheduler`) — запускает `WeatherService.collect_for_station` по расписанию (`UPDATE_INTERVAL_SECONDS`, по умолчанию 3600 секунд), параллелит сбор до N станций.
- `StationRepository` (`app/repositories/station.py`) — читает список активных станций из PostgreSQL.
- `YandexWeatherClient` (`app/clients/yandex_weather.py`) — делает требование к `https://api.weather.yandex.ru/v2/informers?lat=...&lon=...`, обрабатывает `fact`.
- `WeatherService` (`app/services/weather.py`) — валидация, дедупликация, подготовка Kafka-сообщения и публикация, без записи в базу.
- `KafkaPublisher` (`app/core/kafka.py`) — публикует данные в `actual-weather.raw.v1`.
- `RedisClient` (`app/core/redis_client.py`) — хранит метку последнего таймстампа для каждой станции.
- `API` (`app/api/routes`) — health/metrics и CRUD станций остаются для операторского управления, но фактические данные проходят только через Kafka.

## Источник данных и поток

```
Scheduler
    │
    ▼
Получить активные станции (PostgreSQL)
    │
    ▼
API Яндекс.Погоды (fact)
    │
    ▼
Дедупликация в Redis
    │
    ▼
Kafka (topic `actual-weather.raw.v1`)
    │
    ▼
Analytics Service (подписчики)
```

## Переменные окружения (в `.env`)

```
DATABASE_URL=postgresql://postgres:password@postgres:5432/weather?sslmode=disable
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=password
POSTGRES_DB=weather
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC=actual-weather.raw.v1
REDIS_URL=redis://redis:6379/0
YANDEX_WEATHER_API_KEY=replace_me
YANDEX_WEATHER_URL=https://api.weather.yandex.ru/v2/informers
REQUEST_TIMEOUT=10.0
UPDATE_INTERVAL_SECONDS=3600
STATION_NAME_FILTER=
LOG_LEVEL=INFO
```

`UPDATE_INTERVAL_SECONDS` регулирует частоту циклов (секунды). `REQUEST_TIMEOUT` задаёт таймаут HTTP-клиента (секунды). `STATION_NAME_FILTER` позволяет фильтровать станции по имени вместо использования флага `active`. Compose проксиирует внешний порт 8082 на контейнерный 8080.

## Kafka payload

Сообщения публикуются в формате, который читает `core-api`:

```json
{
  "stationId": 1,
  "observedAt": "2026-07-08T06:00:00Z",
  "temperature": 29.3,
  "windSpeed": 12.1,
  "humidity": 71,
  "pressure": 1009,
  "source": "yandex-weather"
}
```

## Логирование и метрики

- Логируются: запуск цикла, количество активных станций, каждый запрос к API Яндекс.Погоды, успешные публикации и ошибки (API/Kafka).
- Метрики Prometheus (`weather_requests_total`, `request_duration_seconds`, `errors_total`, `stations_processed_total`, `http_requests_total`, `http_request_duration_seconds`).
- Все логи пишутся в stdout (structlog JSON) — удобно для ELK.

## Запуск

1. `copy .env.example .env` и заполните `YANDEX_WEATHER_API_KEY`.
2. `python -m pip install -r requirements.txt`.
3. `docker compose up --build` (PostgreSQL, Kafka, Redis, сервис). Или запуск без Docker: `uvicorn app.main:app --host 0.0.0.0 --port ${PORT}`.

## Проверка

- `python -m pytest` — unit-тесты (моки, без Kafka/PostgreSQL).
- `docker compose run --rm -v "${PWD}:/app" actual-weather python -m pytest` — интеграционные тесты внутри контейнера.
