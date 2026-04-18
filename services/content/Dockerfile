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
    CGO_ENABLED=0 GOOS=linux go build -v -o /workspace/bin/content-service ./services/content/cmd/service

FROM debian:bookworm-slim

RUN set -eux; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates; \
    groupadd -r nonroot; \
    useradd --no-log-init -r -g nonroot nonroot; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /workspace/bin/content-service /app/content-service
COPY --from=builder --chown=nonroot:nonroot /workspace/services/content/configs /app/services/content/configs
COPY --from=builder --chown=nonroot:nonroot /workspace/grpc/gen/openapiv2 /app/grpc/gen/openapiv2

ENV CONTENT_CONFIG_PATH=/app/services/content/configs/service.yml

USER nonroot:nonroot

CMD ["/app/content-service"]
