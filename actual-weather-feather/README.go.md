# Actual Weather Service

Микросервис сбора фактических метеоданных для платформы анализа точности прогнозов погоды.

---

## Быстрый старт (для разработчиков)

Требования:

- Go 1.24
- Docker (для локальной инфраструктуры)
- make (опционально) или PowerShell на Windows

Установка зависимостей (Windows PowerShell):

```powershell
cd "C:\Users\semen\OneDrive\Рабочий стол\ssse-final-project\actual-weather-feather"
make deps-windows
# или напрямую
powershell -ExecutionPolicy Bypass -File .\scripts\install_deps.ps1
```

Установка зависимостей (Unix-like):

```sh
cd actual-weather-feather
make deps
```

Сгенерировать `go.sum` / подтянуть модули:

```sh
go mod tidy
```

Запуск локальной инфраструктуры (Postgres, Kafka, Redis) и сервиса через Docker Compose:

```sh
docker compose up --build
```

Запуск сервера напрямую (для разработки):

```sh
go run ./cmd/server
```

Тесты:

```sh
go test ./...
```

Миграции: SQL-файлы находятся в папке `migrations/`. Для локальной БД выполните содержимое `migrations/V1__init.sql` в вашей базе.

## Основные Make цели

- `make deps` — установить зависимости (Unix)
- `make deps-windows` — установить зависимости (Windows PowerShell)
- `make tidy` — `go mod tidy`

## Конфигурация

Приложение настраивается через переменные окружения (по умолчанию указаны значения в `internal/config`):

- `PORT` — порт HTTP-сервера (по умолчанию `8080`)
- `DATABASE_URL` — URL подключения к PostgreSQL
- `KAFKA_BROKERS` — список брокеров Kafka (например `kafka:9092`)
- `REDIS_URL` — URL Redis
- `OPEN_METEO_URL` — базовый URL Open-Meteo API

## HTTP API (минимальный набор)

- `GET /health` — статус сервиса
- `GET /metrics` — Prometheus метрики
- `GET /stations` — список станций (active = true)
- `POST /stations` — создать станцию (JSON `Station`)
- `PUT /stations/{id}` — обновить станцию
- `DELETE /stations/{id}` — удалить станцию

Пример `POST /stations` body:

```json
{
  "name": "Moscow",
  "latitude": 55.75,
  "longitude": 37.61,
  "region": "RU",
  "active": true
}
```

## Тесты и рекомендации

- Тесты находятся в `internal/*` директориях рядом с реализациями (`*_test.go`).
- Для интеграционных тестов рекомендуем поднять docker-compose окружение и использовать тестовую БД.

## Что реализовано в scaffold'e

- Базовая структура проекта (cmd/, internal/, migrations/)
- `internal/config` — загрузка конфигурации из ENV
- `internal/models` — доменные модели `Station` и `WeatherMeasurement`
- `internal/repository` — интерфейсы + `pgx` реализация
- `internal/client/openmeteo` — HTTP-клиент с retry и парсингом hourly/current_weather
- `internal/kafka` — интерфейс `Publisher` и реализация на `kafka-go`
- `internal/service/weather` — бизнес-логика сбора, сохранения и публикации
- `internal/scheduler` — ожидание полного часа, параллельная обработка станций
- `internal/monitoring` — Prometheus метрики
- HTTP-роуты: health и stations CRUD (стабы)

## Следующие шаги (рекомендуется)

1. Запустить `make deps-windows` / `make deps` и `go mod tidy` у себя локально.
2. Запустить `docker compose up --build` и применить миграции к Postgres.
3. Протестировать `GET /health` и `GET /stations`.
4. Доработать валидацию и нормализацию Open-Meteo (если нужны дополнительные поля).
5. Добавить интеграционные тесты для репозитория и scheduler.

Если нужно, могу сейчас:
- добавить подробный `CONTRIBUTING.md` и примеры `curl` запросов;
- завершить CRUD полностью (GET by id, валидация body, 400/404 ответы);
- подготовить `docker-entrypoint` скрипт для автоматического применения миграций при запуске контейнера.


# О проекте

Actual Weather Service является одним из микросервисов системы анализа качества прогнозов погоды.

Его задача — получать фактические погодные данные для набора метеостанций из Open-Meteo API, преобразовывать их в единый внутренний формат, сохранять в базу данных и публиковать события в Kafka для дальнейшей аналитической обработки.

Сам сервис **не занимается сравнением прогноза и фактической погоды**. Его единственная ответственность — предоставить достоверные данные о реальной погоде.

Это соответствует принципу **Single Responsibility Principle (SRP)** и микросервисной архитектуре.

---

# Место сервиса в общей архитектуре

```
                     Forecast Service
                           │
                           ▼
                     Forecast Storage
                           │
                           ▼

Open-Meteo API
       │
       ▼
Actual Weather Service
       │
       ├────────────► PostgreSQL
       │
       └────────────► Kafka
                              │
                              ▼
                 Forecast Accuracy Service
                              │
                              ▼
                        Analytics API
                              │
                              ▼
                            Web UI
```

---

# Основные обязанности сервиса

Сервис отвечает только за следующие задачи:

- получение списка активных станций;
- получение фактической погоды из Open-Meteo;
- валидация ответа;
- нормализация данных;
- сохранение данных в PostgreSQL;
- публикация сообщений в Kafka;
- сбор метрик;
- логирование;
- healthcheck.

Сервис **не выполняет**:

- сравнение прогноза с фактом;
- аналитические запросы;
- расчёт ошибок прогноза;
- построение графиков;
- авторизацию пользователей.

---

# Почему выбран Go

Для реализации сервиса выбран язык **Go**.

Причины выбора:

- высокая производительность;
- лёгкие goroutine позволяют одновременно обрабатывать большое количество станций;
- встроенная поддержка конкурентности;
- небольшое потребление памяти;
- простой деплой одного бинарного файла;
- отличная поддержка Docker и Kubernetes.

Хотя текущая нагрузка относительно небольшая, использование Go позволяет без изменений архитектуры масштабировать сервис при увеличении количества станций или частоты обновления данных.

---

# Технологический стек

| Технология | Назначение |
|------------|------------|
| Go 1.24 | основной язык |
| Chi | HTTP Router |
| PostgreSQL | хранение данных |
| Kafka | очередь сообщений |
| Redis | кэш и дедупликация |
| Prometheus | метрики |
| Docker | контейнеризация |
| Docker Compose | локальная разработка |
| Kubernetes | развёртывание |
| OpenTelemetry | tracing |
| slog / zap | структурированные логи |

---

# Источник данных

Источником фактической погоды является:

https://open-meteo.com/

Используется REST API.

Для каждой станции выполняется запрос примерно следующего вида:

```
GET /v1/forecast

latitude=...

longitude=...

current=temperature_2m,
relative_humidity_2m,
surface_pressure,
wind_speed_10m,
wind_direction_10m,
precipitation
```

Полученный ответ преобразуется во внутреннюю модель приложения.

---

# Поддержка нескольких станций

Сервис должен одновременно обслуживать множество метеостанций.

Все станции хранятся в PostgreSQL.

Каждая запись содержит:

```
Station

id

name

latitude

longitude

region

active
```

Пример:

| id | station | latitude | longitude |
|----|----------|----------|-----------|
| 1 | Moscow | 55.75 | 37.61 |
| 2 | Saint Petersburg | 59.94 | 30.31 |
| 3 | Kazan | 55.79 | 49.12 |

При запуске сервиса список станций считывается из базы данных.

---

# Расписание

Получение данных происходит автоматически.

Интервал:

```
каждый час
```

Например

```
00:00

01:00

02:00

03:00

...

23:00
```

Если сервис был запущен в

```
15:23
```

он ожидает

```
16:00
```

после чего начинает регулярную работу.

---

# Конкурентная обработка

Для каждой станции создаётся отдельная goroutine.

Схема работы:

```
Scheduler

↓

Получить список станций

↓

for station := range stations

↓

go CollectWeather(station)

↓

HTTP Open-Meteo

↓

Validation

↓

Save PostgreSQL

↓

Publish Kafka
```

Таким образом задержка одной станции не влияет на остальные.

---

# Поток обработки данных

```
Open-Meteo API

↓

HTTP Client

↓

Validation

↓

Normalization

↓

Deduplication

↓

PostgreSQL

↓

Kafka
```

---

# PostgreSQL

Используется как операционная база данных.

Основные таблицы:

## stations

```
id

name

latitude

longitude

region

active
```

---

## weather_measurements

```
id

station_id

timestamp

temperature

humidity

pressure

wind_speed

wind_direction

precipitation

created_at
```

---

# Kafka

После успешного сохранения запись публикуется в Kafka.

Топик

```
weather.actual
```

Формат сообщения

```json
{
  "station_id": 1,
  "timestamp": "2026-06-01T15:00:00Z",
  "temperature": 18.6,
  "humidity": 71,
  "pressure": 760,
  "wind_speed": 5.2,
  "wind_direction": 180
}
```

Эти события используются сервисом расчёта точности прогнозов.

---

# Надёжность

Для обеспечения устойчивости используются следующие механизмы.

## Retry

При временной недоступности Open-Meteo или Kafka выполняются повторные попытки.

Используется exponential backoff.

```
1 sec

↓

2 sec

↓

4 sec

↓

8 sec

↓

16 sec
```

---

## Таймауты

Все HTTP-запросы имеют ограничение по времени.

Это предотвращает зависание goroutine.

---

## Дедупликация

Повторно полученные события не должны повторно записываться.

Для этого используется Redis.

---

## Healthcheck

```
GET /health
```

Возвращает состояние:

- PostgreSQL
- Kafka
- Redis
- Open-Meteo (опционально)

---

# Метрики

Метрики экспортируются Prometheus.

Основные:

```
weather_requests_total

weather_request_duration_seconds

stations_processed_total

weather_saved_total

kafka_publish_total

retry_total

errors_total
```

---

# Логирование

Используются структурированные JSON-логи.

Пример

```json
{
  "station_id": 5,
  "duration_ms": 124,
  "status": "success",
  "service": "actual-weather"
}
```

---

# API

## GET /health

Проверка состояния сервиса.

---

## GET /metrics

Метрики Prometheus.

---

## GET /stations

Получение списка станций.

---

## POST /stations

Добавление новой станции.

---

## PUT /stations/{id}

Изменение станции.

---

## DELETE /stations/{id}

Удаление станции.

---

# Структура проекта

```
actual-weather-service/

cmd/
    server/

internal/

    api/

    client/

    config/

    database/

    kafka/

    middleware/

    models/

    monitoring/

    repository/

    service/

migrations/

configs/

Dockerfile

docker-compose.yml

go.mod

README.md
```

---

# Развёртывание

Каждый сервис работает в собственном Docker-контейнере.

В docker-compose поднимаются:

- Actual Weather Service
- PostgreSQL
- Kafka
- Redis

В Kubernetes используются:

- Deployment
- Service
- ConfigMap
- Secret
- Ingress
- HorizontalPodAutoscaler

---

# Масштабирование

Архитектура сервиса позволяет выполнять горизонтальное масштабирование.

При увеличении количества станций достаточно увеличить количество реплик Deployment.

```
replicas: 1

↓

replicas: 3

↓

replicas: 5
```

При этом код приложения изменять не требуется.

Использование goroutine позволяет эффективно использовать ресурсы процессора даже при большом количестве одновременно выполняемых запросов.

---

# Соответствие требованиям проекта

В рамках дипломного проекта сервис реализует следующие требования.

✔ микросервисная архитектура

✔ Docker

✔ Kubernetes

✔ PostgreSQL

✔ Kafka

✔ Redis

✔ REST API

✔ Prometheus

✔ Healthcheck

✔ Retry

✔ Exponential Backoff

✔ Поддержка нескольких станций

✔ Почасовое получение данных

✔ Подготовка данных для аналитического сервиса

---

# Дальнейшее развитие

В дальнейшем сервис может быть доработан без изменения архитектуры.

Возможные направления развития:

- подключение реальных IoT-датчиков вместо Open-Meteo;
- использование MQTT вместо HTTP;
- шардирование по регионам;
- автоматическое обнаружение новых станций;
- кэширование запросов;
- поддержка нескольких поставщиков данных;
- отказоустойчивая работа с несколькими экземплярами Kafka;
- интеграция с системой алертов.

Благодаря разделению ответственности изменения будут затрагивать только слой получения данных (`client`), тогда как бизнес-логика, хранение данных и публикация сообщений останутся неизменными.