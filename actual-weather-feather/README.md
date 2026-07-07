# Actual Weather Service (Python)

Actual Weather Service переписывается на **Python** для выполнения тех же задач, что и исходный Go-сервис:
- собирать фактические погодные данные Open-Meteo;
- сохранять их в PostgreSQL;
- сохранять события в Kafka;
- экспонировать health/metrics и CRUD-интерфейс управления станциями.

> В этом репозитории пока только документация по новому стеку. Реализация должна базироваться на описанной архитектуре и быть написана другим агентом.

---

## Быстрый старт для следующего агента

1. Убедитесь, что у вас установлен Python **3.12+**, Docker и docker-compose.
2. Создайте виртуальное окружение и установите зависимости (в будущем будет `requirements.txt` / `pyproject.toml`).

```bash README.md
python -m venv .venv
source .venv/bin/activate  # или .\.venv\Scripts\Activate.ps1 на Windows
pip install -r requirements.txt  # файл пока пуст, но placeholder должен быть
```

3. Локальная инфраструктура поднимается через существующий docker-compose (Postgres, Kafka, Redis). Обновить сервис, чтобы он использовал новый Python-образ и команды из секции `app/`.

```bash README.md
docker compose up --build
```

4. Основной HTTP-сервер будет запускаться через `uvicorn app.main:app --reload --host 0.0.0.0 --port ${PORT}`.

---

## Архитектура и конвенции

Точный план слоёв и модулей описан в `ARCHITECTURE.md`. Основные различия по сравнению с Go-реализацией:

- **FastAPI** + **uvicorn** для HTTP.
- **httpx.AsyncClient** для Open-Meteo.
- **SQLAlchemy (asyncpg)** или `asyncpg` + шаблон репозитория для PostgreSQL.
- **aiokafka** для публикаций.
- **aioredis** для дедупликации/кэша.
- **prometheus_client** и `FastAPI` middleware для метрик.
- **Pydantic BaseSettings** для загрузки конфигурации.
- **asyncio**-scheduler ожидает следующего полного часа и запускает сбор погодных данных по станции.

Текущая директория `app/` должна разделяться на:

- `app/api` — FastAPI маршруты (health, metrics, stations CRUD).
- `app/services` — бизнес-логика сбора/валидации/нормализации/публикации.
- `app/clients` — Open-Meteo HTTP-клиент.
- `app/repositories` — PostgreSQL-репозитории.
- `app/config` — Pydantic настройки и helpers.
- `app/models` / `app/schemas` — Pydantic модели и ORM‑строки.
- `app/monitoring` — метрики/логирование.
- `app/scheduler` — модуль, дожидающийся полного часа и запускающий сбор ошибок.

Дополнительно:
- `scripts/` — entrypoint (`docker-entrypoint.sh`), setup-утилиты.
- `migrations/` — SQL/Алембик-обновления с исходными таблицами.
- `tests/` — unit/integration для каждого уровня.

---

## Что должен сделать следующий агент

1. **Создать структуру Python-пакетов** (`app/` со слоями выше, `tests/`, `configs/`, `scripts/`).
2. **Описать конфигурацию** через `app/config/settings.py`, поддерживающую `POSTGRES_DSN`, `KAFKA_BROKERS`, `REDIS_URL`, `OPEN_METEO_API`, `PORT`, `LOG_LEVEL`, таймауты и retry-параметры.
3. **Определить модели** `Station` и `WeatherMeasurement` (ORM + Pydantic).
4. **Написать skeleton-репозитории** с методами для CRUD станций и сохранения измерений. Пока оставьте тела `NotImplementedError`, но опишите интерфейсы.
5. **Добавить Open-Meteo клиент** с `httpx.AsyncClient`, retry (exponential backoff), базовой сериализацией.
6. **Склеить scheduler**: `asyncio` задача, которая ждёт следующего полного часа, получает список активных станций и запускает `asyncio.create_task()` для каждой.
7. **Определить отслеживаемые метрики** через `prometheus_client` и middleware.
8. **Описать HTTP-ручки**: `GET /health`, `GET /metrics`, `GET/POST/PUT/DELETE /stations`.
9. **Подготовить `Dockerfile` и `docker-compose.yml`** (ориентируясь на текущие файлы, но заменив Go-исполнение на Python) и добавить `scripts/docker-entrypoint.sh`.
10. **Документировать API** (описать контракт `Station` и формат Kafka-сообщения) прямо в README и `ARCHITECTURE.md`.

---

## Дорожная карта документации

| Раздел | Состояние | Комментарий |
| ------ | --------- | ----------- |
| Архитектура Python (FastAPI + asyncio) | ✅ описана в ARCHITECTURE.md | Дополнить схемы, если появятся новые компоненты |
| README и инструкции | ✅ обновлены | Добавить диаграммы после реализации |
| Реализация | 🚧 **оставлено следующему агенту** | Важно: концентрироваться на async-конкурентности и чистом слоях |

---

## Дополнительно

- В `docker-compose.yml` оставить сервисы Postgres, Kafka, Redis как есть; Python-сервис может именоваться `actual-weather-python`.
- `requirements.txt` пока содержит комментарий о том, что зависимости будут добавлены по мере реализации.
- Все TODO/NotImplemented осталось в документации (не в коде), чтобы снизить объём начальной работы.
- Вся логика должна соблюдаться согласно принципам **Clean Architecture**, **SOLID**, **Dependency Injection** и **context-aware async**.

После создания структуры следующий агент должен вернуть сюда README краткое описание (что реализовано) и отметить `TASK.md` (новую версию) по мере выполнения этапов.
