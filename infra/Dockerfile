# syntax=docker/dockerfile:1.7

FROM golang:1.22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache build-base git sqlite-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ENV CGO_ENABLED=1 GOOS=linux

RUN go build -ldflags="-s -w" -o /app/server ./cmd/server

FROM alpine:3.20
WORKDIR /app

RUN adduser -D -H appuser && \
    apk add --no-cache ca-certificates sqlite-libs

COPY --from=builder /app/server /app/server

RUN mkdir -p /app/data && chown -R appuser:appuser /app
USER appuser

ENV APP_ENV=dev \
    APP_PORT=8080 \
    DB_PATH=/app/data/app.db \
    JWT_SECRET=dev-secret-change-me \
    CORS_ORIGINS=* \
    LOG_LEVEL=info \
    RATE_LIMIT_RPS=10 \
    RATE_LIMIT_BURST=20

EXPOSE 8080

ENTRYPOINT ["/app/server"]


