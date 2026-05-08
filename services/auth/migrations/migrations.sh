#!/bin/sh
set -eu

GOOSE_DBSTRING="host=${POSTGRES_HOST} port=${POSTGRES_PORT} user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} database=${POSTGRES_DBNAME}"
if [ "${SSL_MODE:-false}" = true ]; then
  GOOSE_DBSTRING="${GOOSE_DBSTRING} sslmode=require"
else
  GOOSE_DBSTRING="${GOOSE_DBSTRING} sslmode=disable"
fi

export GOOSE_DBSTRING
export GOOSE_DRIVER=postgres

goose -v -dir /opt/microservice/migrations/goose up
