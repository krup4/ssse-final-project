# Зависимости проекта

Этот файл перечисляет внешние модули, используемые в сервисе, и команды для их установки.

Команды для установки (POSIX):

```sh
make deps
```

Команды для установки (Windows PowerShell):

```ps1
make deps-windows
```

Список модулей:

- github.com/go-chi/chi/v5 (HTTP router)
- github.com/jackc/pgx/v5 (PostgreSQL)
- github.com/segmentio/kafka-go (Kafka producer)
- github.com/redis/go-redis/v9 (Redis client)
- github.com/prometheus/client_golang (Prometheus metrics)
- github.com/cenkalti/backoff/v4 (exponential backoff)

После установки выполните `go mod tidy` (Makefile содержит цель `tidy`).
