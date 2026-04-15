FROM golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /app/server ./cmd/app

FROM debian:bookworm-slim

RUN set -eux; \
    apt-get update; \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates; \
    groupadd -r nonroot; \
    useradd --no-log-init -r -g nonroot nonroot; \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder --chown=nonroot:nonroot /app/server /app/server
COPY --from=builder --chown=nonroot:nonroot /app/config.yml /app/config.yml

USER nonroot:nonroot

CMD ["/app/server"]
