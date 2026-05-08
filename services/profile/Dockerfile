# syntax=docker/dockerfile:1

ARG GOOSE_VERSION_TAG="v3.27.0"

FROM golang:1.25-bookworm AS build

ENV GOCACHE=/tmp/go_cache

WORKDIR /src

COPY go.mod go.sum ./

RUN --mount=type=cache,target=/tmp/go_cache/ \
    go mod download

COPY ./ ./

RUN --mount=type=cache,target=/tmp/go_cache/ \
    CGO_ENABLED=0 GOOS=linux go build -o /out/profile-service ./services/profile/cmd/service

COPY ./services/profile/configs/ /out/configs/
COPY ./services/profile/migrations/goose/ /out/migrations/goose/
COPY --chmod=755 ./services/profile/migrations/migrations.sh /out/migrations.sh
COPY ./grpc/gen/openapiv2/ /out/grpc/gen/openapiv2/

FROM gcr.io/distroless/static-debian12 AS build-release-stage

WORKDIR /opt/microservice

COPY --from=build /out/ ./

ARG GOOSE_VERSION_TAG
ADD --chmod=755 "https://github.com/pressly/goose/releases/download/$GOOSE_VERSION_TAG/goose_linux_x86_64" /bin/goose
COPY --from=busybox:1.37.0-uclibc /bin/sh /bin/sh

ENV PROFILE_CONFIG_PATH=/opt/microservice/configs/service.yml

USER nonroot:nonroot

EXPOSE 8082 9092

ENTRYPOINT ["./profile-service"]
