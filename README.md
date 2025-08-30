# web-analyzer-go

## Overview
A simple web page analyzer built with Go. It exposes a REST API and a minimal web UI that accepts a URL and returns:
- HTML version
- Page title
- Heading counts (h1–h6)
- Link statistics (internal, external, inaccessible)
- Presence of a login form

The app also provides health, metrics, profiling, structured logging, and graceful shutdown.

## Technology Stack
- **Backend**: Go `net/http`, `golang.org/x/net/html` (HTML parsing), structured logging via `log/slog`, graceful shutdown, concurrent analysis with goroutines.
- **Frontend**: Static page served from `web/` using React 18 (UMD) and Tailwind CSS via CDN.
- **API Docs**: Swagger (served at runtime) using `swaggo/http-swagger` and static specs in `docs/swagger.yaml` and `docs/swagger.json`.
- **DevOps/Observability**: Docker multi-stage build, Prometheus metrics (`promhttp`), pprof (`/debug/pprof/`).

## URLs (local)
- **Backend base URL (native run)**: `http://localhost:8080`
- **Frontend UI**: `http://localhost:8080/`
- **Health**: `http://localhost:8080/health`
- **Metrics (Prometheus format)**: `http://localhost:8080/metrics`
- **Profiling (pprof)**: `http://localhost:8080/debug/pprof/`
- **API Docs (Swagger UI)**: `http://localhost:8080/swagger/` (or `.../swagger/index.html`)

Docker run note: the server listens on port 8080 inside the container. Map to any host port as needed, e.g. `-p 8080:8080` to access at `http://localhost:8080`.

## Prerequisites
- Go 1.23+
- Docker (optional)

## Setup and Run
1) Native run
```sh
go mod tidy
go run ./cmd/main.go
# App at http://localhost:8080
```

2) Docker
```sh
docker build -t web-analyzer-go .
docker run --rm -p 8080:8080 web-analyzer-go
# App at http://localhost:8080
```

## API Usage
- `POST /analyze`
  - Request body:
    ```json
    { "url": "https://simplewebapp.com" }
    ```
  - Success response (200):
    ```json
    {
      "html_version": "HTML5",
      "title": "Example Domain",
      "headings": [ { "level": 1, "count": 1 }, ... ],
      "links": { "internal": 3, "external": 2, "inaccessible": 1 },
      "login_form": false
    }
    ```
  - Error responses: 400 (invalid input), 502 (upstream/unreachable)

- Supporting endpoints
  - `GET /health` — Health check
  - `GET /metrics` — Prometheus metrics
  - `GET /debug/pprof/` — Profiling
  - `GET /swagger/` — Swagger UI

### cURL examples
```sh
curl -s http://localhost:8080/health

curl -s -X POST http://localhost:8080/analyze \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://simplewebapp.com"}'
```

## Project Structure
- `cmd/` — Main entrypoint
- `internal/api/` — HTTP handlers and router
- `internal/service/` — Analysis orchestration and facade
- `internal/analyzer/` — Analyzer core, strategies, and execution
- `internal/model/` — DTOs / response models
- `internal/middleware/` — Request ID, recoverer, and structured logging
- `internal/metrics/` — Metrics integration
- `internal/util/` — Logging setup and HTML/link utilities
- `internal/factory/` — HTTP client and link checker factory
- `docs/` — Swagger specs and generated docs
- `web/` — Static frontend

## Potential improvements

### Performance
- Bounded per-host concurrency for link checks to avoid overloading single domains.
- TTL LRU cache for link accessibility and HEAD-not-supported hosts to reduce duplicate upstream calls.
- Switch some checks to `html.Tokenizer` to avoid full-tree parsing on very large documents.
- Adaptive timeouts and retry-with-backoff policy for link checks; short deadlines for HEAD, longer for GET fallback.

### Architecture
- Pass `context.Context` into strategies and `LinkChecker` (`Analyze(ctx, ...)`, `IsAccessible(ctx, ...)`) for true cancellation.
- Strategy plugin system (DI/registration) with clear interfaces and isolated packages for easier extension/testing.
- Centralized config (env/flags) for pool sizes, timeouts, limits; injectable via factory to enable per-request overrides.
- Add tracing (OpenTelemetry) to correlate HTTP fetch, parsing, and each strategy/link-check span across requests.

## Limitations
- Dynamic sites that build HTML with JavaScript (SPAs) are not fully analyzed. The analyzer fetches the initial HTML only and does not execute JS. To support these sites, add an optional headless rendering step (e.g., chromedp/Playwright) or call an external render service.

