Write-Host "Installing Go module dependencies..."

go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go get github.com/segmentio/kafka-go
go get github.com/redis/go-redis/v9
go get github.com/prometheus/client_golang
go get github.com/cenkalti/backoff/v4

go mod tidy

Write-Host "Dependencies installed. Run 'make tidy' if needed."
