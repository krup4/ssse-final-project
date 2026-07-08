FROM golang:1.25-alpine3.22 AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/core-api ./cmd/core-api

FROM scratch AS runner

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /out/core-api /app/core-api

USER 10001:10001

EXPOSE 8080

CMD ["/app/core-api"]
