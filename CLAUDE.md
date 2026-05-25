# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code Style

**No comments.** Do not write any comments in any file — `.go`, `.sql`, `.proto`, or any other. Code must be self-explanatory through naming and structure alone.

## Project Overview

Go microservices backend for **SPORT.tech** — a sports content platform (sporteon.ru). Four services communicate via gRPC, with an HTTP REST layer exposed through grpc-gateway.

## Commands

```bash
# Lint
make lint            # runs gofmt check + go vet

# Test
make test            # go test ./...
go test ./services/auth/...                          # single service
go test ./services/auth/internal/usecase/...         # single package
go test -run TestServiceName ./path/to/package/...   # single test

# Coverage
make coverage        # generate coverage.out report
make coverage-check  # enforce 60% minimum (runs in CI)
make coverage-html   # open HTML report

# CI (what runs on every PR)
make ci              # lint + test + coverage-check

# Code generation
make generate        # proto + easyjson
make proto           # regenerate from grpc/proto/**/*.proto
make easyjson        # regenerate easyjson for content/stripe

# Docker
make compose-up      # build & start full stack (requires .env)
make compose-down
make docker-build    # build all 4 service images
```

## Architecture

### Service Layout

```
services/
  api-gateway/   — HTTP REST → gRPC proxy (port 8080)
  auth/          — Authentication & sessions (gRPC :9091, HTTP :8081)
  profile/       — User profiles, avatars, sport types (gRPC :9092, HTTP :8082)
  content/       — Posts, subscriptions, donations, payments (gRPC :9093, HTTP :8083)
grpc/
  proto/         — .proto source files (auth, profile, content, gateway)
  gen/go/        — generated Go code (DO NOT EDIT)
  gen/openapiv2/ — generated OpenAPI specs
```

### Internal Package Structure (same in every service)

Each service follows clean/hexagonal architecture:

- **`domain/`** — pure business types, no dependencies
- **`usecase/`** — business logic; `ports.go` defines interfaces (repositories, external clients); `service.go` implements them
- **`adapters/`**
  - `grpc/` — gRPC server handler wiring use-case interfaces
  - `repository/postgres/` — PostgreSQL implementations of repository ports
  - `client/` — external service clients (MinIO, Stripe, crypto)
  - `mappers/` — proto ↔ domain conversions
- **`infrastructure/`**
  - `bootstrap/app.go` — wires everything together, owns server lifecycle
  - `config/config.go` — cleanenv config (YAML + env var override)
  - `grpcserver/` — gRPC server setup with interceptors/metrics
  - `httpgateway/` — grpc-gateway HTTP mux (auth/profile/content have local mux; api-gateway aggregates all)
  - `metrics/` — Prometheus metrics
  - `health/` — `/healthz` endpoint

### api-gateway Role

The `api-gateway` is the only service exposed to the internet. It:
1. Runs a grpc-gateway HTTP mux that registers all domain services' handlers
2. Connects to auth/profile/content services via gRPC for actual execution
3. Handles CORS, CSRF, session cookie → gRPC metadata propagation (`x-session-token`), request logging, and Prometheus metrics
4. Serves the merged OpenAPI spec at `/openapi/gateway.swagger.json`

Each microservice also runs its own local grpc-gateway HTTP mux (for direct access / health / OpenAPI).

### Configuration

Config is loaded from `services/<name>/configs/service.yml` then overridden by environment variables. The `config.go` pattern per service uses `cleanenv` with struct tags `yaml:"..."` and `env:"SERVICE_VARNAME"`. Required env vars for docker compose are in `.env` (not committed).

### Mocks

Mocks are hand-written structs in `internal/mocks/` with `*Func` fields (not generated). Pattern:
```go
mock := mocks.AuthUseCase{
    RegisterFunc: func(...) { ... },
}
```

### Databases

Three separate PostgreSQL databases (`sporttech_auth`, `sporttech_profile`, `sporttech_content`). Migrations are shell scripts at `services/<name>/migrations/migrations.sh`. Each service runs a dedicated DB user with minimal grants (set up by `*-db-security` compose services).

### Object Storage

MinIO for avatars (profile service) and post media (content service). Public base URL is configurable via `*_STORAGE_PUBLIC_BASE_URL` env vars.

### Payments

Content service integrates Stripe (`STRIPE_SECRET_KEY` required). The `payment_provider.go` uses easyjson for serialization.
