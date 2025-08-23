# test-project-go

## Overview
Enterprise-grade Go API boilerplate with layered monolith architecture, observability, and best practices. No database required.

## Features
- Layered architecture (API, Service, Utilities)
- Health check endpoint
- Prometheus metrics (`/metrics`)
- pprof profiling (`/debug/pprof/`)
- Structured logging (slog)
- Graceful shutdown
- Multi-stage Dockerfile

## Getting Started

### Prerequisites
- Go 1.21+
- Docker (optional)

### Setup
```sh
go mod tidy
go run ./cmd/main.go
```

### Build & Run with Docker
```sh
docker build -t test-project-go .
docker run -p 8080:8080 test-project-go
```

## API Endpoints
- `GET /health` — Health check
- `GET /metrics` — Prometheus metrics
- `GET /debug/pprof/` — pprof profiling

## Testing
```sh
go test ./...
```

## Project Structure
- `cmd/` — Main entrypoint
- `internal/api/` — HTTP handlers
- `internal/service/` — Business logic
- `internal/util/` — Utilities (logging)
- `internal/metrics/` — Metrics setup

## License
MIT
