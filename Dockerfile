# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.24
ARG ALPINE_VERSION=3.20

FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --no-cache git ca-certificates
WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/webhook-server ./cmd/webhook-server/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/frontol-loader ./cmd/loader/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/frontol-loader-local ./cmd/loader-local/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/parser-test ./cmd/parser-test/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/send-request ./cmd/send-request/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/clear-requests ./cmd/clear-requests/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/clear-db ./cmd/clear-db/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/ftp-server ./cmd/ftp-server/main.go && \
    CGO_ENABLED=0 GOOS=linux \
    go build -trimpath -ldflags="-s -w" -o /out/ftp-check ./cmd/ftp-check/main.go

FROM alpine:${ALPINE_VERSION}

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1000 -S ftpgroup && \
    adduser -u 1000 -S -G ftpgroup -s /sbin/nologin ftpuser && \
    addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S -G appgroup -s /sbin/nologin appuser

WORKDIR /app

COPY --from=builder /out/webhook-server /app/webhook-server
COPY --from=builder /out/frontol-loader /app/frontol-loader
COPY --from=builder /out/migrate /app/migrate
COPY --from=builder /out/frontol-loader-local /app/frontol-loader-local
COPY --from=builder /out/parser-test /app/parser-test
COPY --from=builder /out/send-request /app/send-request
COPY --from=builder /out/clear-requests /app/clear-requests
COPY --from=builder /out/clear-db /app/clear-db
COPY --from=builder /out/ftp-server /app/ftp-server
COPY --from=builder /out/ftp-check /app/ftp-check
COPY --from=builder /src/pkg/migrate/migrations /app/migrations

COPY --chown=appuser:appgroup scripts/clear-database.sql /app/scripts/clear-database.sql
COPY --chown=appuser:appgroup api/openapi.yaml /app/api/openapi.yaml

RUN install -d -o appuser -g appgroup /app/tmp/frontol

USER appuser
EXPOSE 8080
CMD ["./webhook-server"]
