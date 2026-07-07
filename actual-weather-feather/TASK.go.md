# Actual Weather Service Development Roadmap

Этот документ используется как список задач проекта.

После выполнения задачи необходимо:

- отметить её как выполненную `[x]`;
- при необходимости добавить комментарий;
- не удалять выполненные задачи;
- новые задачи добавлять в соответствующий раздел.

---

# Project Status

**Current Status**

```
Planning
```

---

# Phase 0 — Project Initialization

## Repository

- [ ] Создать Go-модуль
- [ ] Создать структуру проекта
- [ ] Создать README.md
- [ ] Создать AI.md
- [ ] Создать TASK.md
- [ ] Настроить .gitignore

---

## Configuration

- [ ] Реализовать загрузку конфигурации
- [ ] Поддержка ENV
- [ ] Конфигурация PostgreSQL
- [ ] Конфигурация Kafka
- [ ] Конфигурация Redis
- [ ] Конфигурация Open-Meteo
- [ ] Конфигурация Scheduler

---

# Phase 1 — Infrastructure

## Docker

- [ ] Dockerfile
- [ ] docker-compose.yml
- [ ] Volume для PostgreSQL
- [ ] Healthcheck контейнеров

---

## PostgreSQL

- [ ] Подключение pgx
- [ ] Создать миграции
- [ ] Таблица stations
- [ ] Таблица weather_measurements
- [ ] Индексы
- [ ] Проверить подключение

---

## Redis

- [ ] Подключение
- [ ] Проверка соединения

---

## Kafka

- [ ] Подключение Producer
- [ ] Создать topic weather.actual
- [ ] Проверка публикации

---

# Phase 2 — Domain Models

## Models

- [ ] Station
- [ ] WeatherMeasurement
- [ ] OpenMeteoResponse

---

## Repository

### Station Repository

- [ ] GetAll()
- [ ] GetByID()
- [ ] Create()
- [ ] Update()
- [ ] Delete()

---

### Weather Repository

- [ ] Save()
- [ ] GetLatest()
- [ ] GetByStation()

---

# Phase 3 — Open-Meteo Client

## HTTP Client

- [ ] Создать http.Client
- [ ] Настроить Timeout
- [ ] Настроить Retry
- [ ] Реализовать Exponential Backoff
- [ ] Использовать context.Context

---

## Open-Meteo API

- [ ] Получение текущей погоды
- [ ] Парсинг JSON
- [ ] Проверка ошибок API
- [ ] Нормализация данных

---

# Phase 4 — Scheduler

## Scheduler

- [ ] Запуск каждый полный час
- [ ] Корректная работа после рестарта
- [ ] Graceful shutdown

---

## Station Processing

- [ ] Загрузка списка станций
- [ ] Обработка только active=true
- [ ] Goroutine на каждую станцию
- [ ] WaitGroup
- [ ] Context cancellation

---

# Phase 5 — Business Logic

## Validation

- [ ] Проверка координат
- [ ] Проверка температуры
- [ ] Проверка влажности
- [ ] Проверка давления
- [ ] Проверка скорости ветра

---

## Normalization

- [ ] Timestamp
- [ ] Единицы измерения
- [ ] Проверка пустых значений

---

## Deduplication

- [ ] Redis cache
- [ ] Проверка повторных сообщений

---

# Phase 6 — Persistence

## PostgreSQL

- [ ] Сохранение измерения
- [ ] Обработка ошибок
- [ ] Retry

---

## Kafka

- [ ] JSON serialization
- [ ] Publish weather.actual
- [ ] Retry
- [ ] Подтверждение публикации

---

# Phase 7 — REST API

## Health

- [ ] GET /health

---

## Metrics

- [ ] GET /metrics

---

## Stations

- [ ] GET /stations
- [ ] POST /stations
- [ ] PUT /stations/{id}
- [ ] DELETE /stations/{id}

---

# Phase 8 — Monitoring

## Logging

- [ ] slog
- [ ] Structured logs
- [ ] Request ID
- [ ] Error logging

---

## Prometheus

- [ ] weather_requests_total
- [ ] weather_saved_total
- [ ] stations_processed_total
- [ ] retry_total
- [ ] request_duration_seconds
- [ ] errors_total

---

# Phase 9 — Reliability

## Retry

- [ ] Open-Meteo
- [ ] PostgreSQL
- [ ] Kafka
- [ ] Redis

---

## Timeouts

- [ ] HTTP
- [ ] PostgreSQL
- [ ] Kafka

---

## Graceful Shutdown

- [ ] Context cancellation
- [ ] Закрытие Kafka Producer
- [ ] Закрытие PostgreSQL
- [ ] Закрытие Redis

---

# Phase 10 — Testing

## Unit Tests

- [ ] Repository
- [ ] Service
- [ ] Scheduler
- [ ] Open-Meteo Client

---

## Integration Tests

- [ ] PostgreSQL
- [ ] Kafka
- [ ] Redis

---

## HTTP Tests

- [ ] Health
- [ ] Stations API

---

## Concurrency Tests

- [ ] Race detector
- [ ] Parallel station processing

---

# Phase 11 — Deployment

## Docker

- [ ] Multi-stage build
- [ ] Минимальный образ
- [ ] Проверка запуска

---

## Kubernetes

- [ ] Deployment
- [ ] Service
- [ ] ConfigMap
- [ ] Secret
- [ ] Ingress
- [ ] HPA

---

# Phase 12 — Documentation

- [ ] README актуализирован
- [ ] AI.md актуализирован
- [ ] Комментарии в публичных методах
- [ ] Swagger/OpenAPI (если потребуется)

---

# Quality Checklist

Перед завершением проекта необходимо убедиться, что:

- [ ] Проект компилируется
- [ ] go fmt выполнен
- [ ] go vet выполнен
- [ ] go test проходит
- [ ] go test -race проходит
- [ ] Нет panic
- [ ] Нет data race
- [ ] Все ошибки обработаны
- [ ] Нет TODO в коде
- [ ] Нет закомментированного кода
- [ ] Все зависимости внедряются через конструкторы
- [ ] Соблюдена Clean Architecture

---

# Production Improvements (Future)

Не входят в диплом, но могут быть реализованы позже.

- [ ] MQTT вместо HTTP
- [ ] Поддержка нескольких погодных провайдеров
- [ ] Circuit Breaker
- [ ] Distributed Tracing
- [ ] Rate Limiter
- [ ] Local Queue
- [ ] ClickHouse Writer
- [ ] S3 Archive
- [ ] CI/CD Pipeline
- [ ] Helm Chart
- [ ] Grafana Dashboard

---

# Current Sprint

Следующая задача для AI-агента:

> Начать с **Phase 0**, создать структуру проекта, Go-модуль, Docker Compose и базовую конфигурацию. После успешного завершения отметить соответствующие пункты как выполненные и перейти к Phase 1.

---

# Notes

В процессе разработки придерживаться следующих принципов:

- Clean Architecture
- SOLID
- DRY
- KISS
- Repository Pattern
- Dependency Injection
- Context-first подход
- Graceful Shutdown
- Structured Logging
- Horizontal Scalability

При возникновении неоднозначностей выбирать решение, наиболее близкое к production-практикам.