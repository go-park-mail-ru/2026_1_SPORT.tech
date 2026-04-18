# syntax=docker/dockerfile:1.7

FROM golang:1.25-bookworm AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    update-ca-certificates && \
    go mod download

COPY . ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -v -o /workspace/bin/auth-service ./services/auth/cmd/service

FROM debian:bookworm-slim

RUN set -eux; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates; \
    groupadd -r nonroot; \
    useradd --no-log-init -r -g nonroot nonroot; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /workspace/bin/auth-service /app/auth-service
COPY --from=builder --chown=nonroot:nonroot /workspace/services/auth/configs /app/services/auth/configs
COPY --from=builder --chown=nonroot:nonroot /workspace/grpc/gen/openapiv2 /app/grpc/gen/openapiv2

ENV AUTH_CONFIG_PATH=/app/services/auth/configs/service.yml

USER nonroot:nonroot

CMD ["/app/auth-service"]
