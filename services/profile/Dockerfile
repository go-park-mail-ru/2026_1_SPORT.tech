FROM golang:1.25-bookworm AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /workspace/bin/profile-service ./services/profile/cmd/service

FROM debian:bookworm-slim

RUN set -eux; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends ca-certificates; \
    groupadd -r nonroot; \
    useradd --no-log-init -r -g nonroot nonroot; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /workspace/bin/profile-service /app/profile-service
COPY --from=builder --chown=nonroot:nonroot /workspace/services/profile/configs /app/services/profile/configs
COPY --from=builder --chown=nonroot:nonroot /workspace/grpc/gen/openapiv2 /app/grpc/gen/openapiv2

ENV PROFILE_CONFIG_PATH=/app/services/profile/configs/service.yml

USER nonroot:nonroot

CMD ["/app/profile-service"]
