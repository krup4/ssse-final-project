FROM golang:1.25-alpine3.22 AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/core-api ./cmd/core-api

FROM alpine:3.22 AS runner

RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 app

WORKDIR /app

COPY --from=builder /out/core-api /app/core-api

USER app

EXPOSE 8080

CMD ["/app/core-api"]
