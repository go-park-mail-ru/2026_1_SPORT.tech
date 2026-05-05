# syntax=docker/dockerfile:1

FROM golang:1.25-bookworm AS build

ENV GOCACHE=/tmp/go_cache

WORKDIR /src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/tmp/go_cache/ \
    go mod download

COPY ./ ./

RUN --mount=type=cache,target=/tmp/go_cache/ \
    CGO_ENABLED=0 GOOS=linux go build -o /out/auth-service ./services/auth/cmd/service

COPY ./services/auth/configs/ /out/configs/
COPY ./grpc/gen/openapiv2/ /out/grpc/gen/openapiv2/

FROM gcr.io/distroless/static-debian12 AS build-release-stage

WORKDIR /opt/microservice

COPY --from=build /out/ ./

ENV AUTH_CONFIG_PATH=/opt/microservice/configs/service.yml

USER nonroot:nonroot

EXPOSE 8081 9091

ENTRYPOINT ["./auth-service"]
