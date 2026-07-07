FROM golang:1.25-alpine3.22 AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /opt
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN go build -o bin/application .

FROM alpine:3.22 AS runner

RUN apk update && apk add ca-certificates libc6-compat openssh bash && rm -rf /var/cache/apk/*

WORKDIR /opt

COPY . . 

COPY --from=builder /opt/bin/application ./

CMD ["bin/application"]