## Общая информация

Actual Weather Service — микросервис, отвечающий за получение фактических погодных данных для платформы анализа точности прогнозов.

Сервис периодически получает данные из Open-Meteo API, нормализует их, сохраняет в PostgreSQL и публикует события в Kafka для последующей аналитической обработки.

---

# Архитектура системы

```text
                        +-------------------------+
                        |     Open-Meteo API      |
                        +-----------+-------------+
                                    |
                              HTTPS Requests
                                    |
                                    ▼
                      +-----------------------------+
                      | Actual Weather Service      |
                      |-----------------------------|
                      | Scheduler                   |
                      | Weather Service             |
                      | Validation                  |
                      | Repository                  |
                      | Kafka Publisher             |
                      +--------------+--------------+
                                     |
                  +------------------+------------------+
                  |                                     |
                  ▼                                     ▼
         PostgreSQL                            Kafka Topic
     weather_measurements                  weather.actual
                  |                                     |
                  +------------------+------------------+
                                     |
                                     ▼
                   Forecast Accuracy Service
                                     |
                                     ▼
                            Analytics API
                                     |
                                     ▼
                                  Web UI
```

---

# Основные компоненты

## Scheduler

Запускает получение погодных данных каждый полный час.

Пример:

```
00:00

01:00

02:00

...

23:00
```

Scheduler отвечает только за запуск обработки.

Он не знает ничего о PostgreSQL, Kafka или HTTP.

---

## Weather Service

Основная бизнес-логика приложения.

Последовательность работы:

```
Получить станции

↓

Для каждой станции

↓

Получить погоду

↓

Провалидировать

↓

Нормализовать

↓

Сохранить

↓

Опубликовать Kafka Event
```

Вся бизнес-логика должна находиться только здесь.

---

## Open-Meteo Client

Отвечает исключительно за взаимодействие с Open-Meteo API.

Функции клиента:

- выполнение HTTP-запросов;
- обработка ошибок;
- retry;
- десериализация JSON;
- возврат доменной модели.

Клиент ничего не знает про базу данных.

---

## Repository

Repository реализует доступ к PostgreSQL.

Обязанности:

- чтение списка станций;
- сохранение измерений;
- получение последних измерений.

Repository не содержит бизнес-логики.

---

## Kafka Publisher

После успешного сохранения данных публикует событие в Kafka.

Topic:

```
weather.actual
```

Формат сообщений — JSON.

---

# Диаграмма обработки одной станции

```text
Station

↓

Open-Meteo Client

↓

HTTP Request

↓

JSON Response

↓

Validation

↓

Normalization

↓

Repository

↓

PostgreSQL

↓

Kafka Publisher

↓

weather.actual
```

---

# Конкурентная обработка

Получение данных должно происходить одновременно для всех станций.

Используются goroutines.

```text
Load Stations

        │

        ▼

+------------------------------+

Station 1  → goroutine

Station 2  → goroutine

Station 3  → goroutine

Station N  → goroutine

+------------------------------+

        │

        ▼

WaitGroup

        │

        ▼

Finish
```

Преимущества:

- независимость станций;
- высокая производительность;
- лёгкое масштабирование.

---

# Scheduler

Алгоритм работы.

```text
Service Start

↓

Load Config

↓

Connect Database

↓

Wait until next full hour

↓

Load Active Stations

↓

Start Goroutines

↓

Collect Weather

↓

Sleep until next full hour

↓

Repeat
```

Важно:

Получение данных происходит именно в моменты:

```
12:00

13:00

14:00
```

а не через каждые 60 минут после запуска.

---

# Sequence Diagram

```text
Scheduler

 |

 | Load Stations

 ▼

Repository

 |

 | SELECT stations

 ▼

Scheduler

 |

 | for station

 ▼

Weather Service

 |

 | HTTP Request

 ▼

Open-Meteo API

 |

 | JSON

 ▼

Weather Service

 |

 | Validate

 |

 | Normalize

 ▼

Repository

 |

 | INSERT weather_measurements

 ▼

PostgreSQL

 |

 ▼

Weather Service

 |

 | Publish

 ▼

Kafka

 |

 ▼

weather.actual
```

---

# Архитектура проекта

```text
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

configs/

migrations/
```

---

# Ответственность пакетов

## api

REST API.

Содержит:

- Health
- Metrics
- CRUD станций

---

## client

Работа с Open-Meteo.

Не знает ничего о PostgreSQL.

---

## service

Бизнес-логика.

Оркестрирует остальные компоненты.

---

## repository

Работа с PostgreSQL.

Не содержит HTTP.

---

## kafka

Публикация событий.

---

## monitoring

Prometheus.

Health.

Tracing.

Logging.

---

## config

Конфигурация приложения.

---

# Поток данных

```text
Open-Meteo

↓

HTTP

↓

Client

↓

Service

↓

Validation

↓

Repository

↓

PostgreSQL

↓

Kafka

↓

Forecast Accuracy Service
```

---

# База данных

## stations

```text
id

name

latitude

longitude

region

active
```

---

## weather_measurements

```text
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

# Масштабирование

Архитектура допускает горизонтальное масштабирование.

```text
                 Kubernetes

          +----------------------+

           Replica 1

           Replica 2

           Replica 3

          +----------------------+

                    │

                    ▼

               PostgreSQL

                    │

                    ▼

                 Kafka
```

При увеличении числа станций достаточно увеличить количество реплик Deployment.

Изменения кода не требуются.

---

# Отказоустойчивость

Используются следующие механизмы.

## Retry

Для:

- Open-Meteo
- PostgreSQL
- Kafka

Exponential Backoff:

```
1s

↓

2s

↓

4s

↓

8s

↓

16s
```

---

## Timeout

Каждый HTTP-запрос ограничен по времени.

---

## Graceful Shutdown

При завершении работы:

- завершаются goroutines;
- закрываются соединения;
- завершается Kafka Producer;
- закрывается PostgreSQL.

---

# Метрики

Prometheus экспортирует:

```text
weather_requests_total

weather_saved_total

weather_request_duration_seconds

stations_processed_total

retry_total

errors_total
```

---

# Логирование

Используются структурированные JSON-логи.

Пример:

```json
{
  "service": "actual-weather",
  "station_id": 15,
  "duration_ms": 128,
  "status": "success"
}
```

---

# Основные принципы

Архитектура строится на следующих принципах:

- Clean Architecture
- SOLID
- Repository Pattern
- Dependency Injection
- Context-first
- Horizontal Scalability
- Graceful Shutdown

Каждый пакет отвечает только за одну область ответственности.

Все зависимости внедряются через конструкторы.

Бизнес-логика сосредоточена исключительно в пакете `service`.