#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
ENV_FILE="perf_test/.env.perf"
RESULTS_DIR="perf_test/results"
HOST="${HOST:-http://localhost:8083}"

TARGET="${TARGET:-100000}"
CREATE_RATE="${CREATE_RATE:-400}"
READ_RATE="${READ_RATE:-500}"
READ_DURATION="${READ_DURATION:-30s}"
READ_COUNT="${READ_COUNT:-20000}"

DB_USER="$(grep -E '^DB_USER=' "$ENV_FILE" | cut -d= -f2)"

compose() { docker compose --env-file "$ENV_FILE" "$@"; }

psql_content() {
  compose exec -T db psql -U "$DB_USER" -d sporttech_content -tAc "$1"
}

post_count() { psql_content "SELECT count(*) FROM content_post;" | tr -d '[:space:]'; }

wait_for_service() {
  echo ">> ждём content-service на $HOST ..."
  for _ in $(seq 1 60); do
    if curl -fsS -o /dev/null "$HOST/v1/posts/1" 2>/dev/null \
       || curl -sS -o /dev/null -w '%{http_code}' "$HOST/v1/posts/1" 2>/dev/null | grep -qE '200|404'; then
      echo ">> сервис отвечает"
      return 0
    fi
    sleep 2
  done
  echo "!! сервис не поднялся" >&2
  return 1
}

cmd_up() {
  compose up -d --build db content-service
  wait_for_service
}

cmd_gen() {
  echo ">> генерация таргетов ..."
  python3 perf_test/gen_targets.py \
    --count "$TARGET" --read-count "$READ_COUNT" \
    --max-id "$TARGET" --host "$HOST" --out-dir perf_test
}

cmd_create() {
  mkdir -p "$RESULTS_DIR"
  wait_for_service
  [ -f perf_test/create_targets.json ] || cmd_gen

  local before; before="$(post_count)"
  echo ">> постов в базе до теста: $before"

  local iter=0
  while :; do
    local current remaining
    current="$(post_count)"
    remaining=$(( TARGET - current ))
    if [ "$remaining" -le 0 ]; then break; fi
    iter=$(( iter + 1 ))
    if [ "$iter" -gt 6 ]; then
      echo "!! слишком много итераций, останавливаемся (создано $current)"; break
    fi

    local duration; duration=$(( (remaining + CREATE_RATE - 1) / CREATE_RATE ))
    echo ">> итерация $iter: нужно ещё $remaining, attack ${CREATE_RATE}/s в течение ${duration}s ..."
    vegeta attack \
      -targets=perf_test/create_targets.json -format=json -lazy \
      -rate="$CREATE_RATE" -duration="${duration}s" -timeout=10s \
      > "$RESULTS_DIR/create.bin"
    vegeta report "$RESULTS_DIR/create.bin" | tee "$RESULTS_DIR/create_report_iter${iter}.txt"
  done

  echo ">> постов в базе после теста: $(post_count)"
  echo ">> итоговый отчёт по последней итерации создания:"
  vegeta report "$RESULTS_DIR/create.bin" | tee "$RESULTS_DIR/create_report.txt"
  vegeta report -type=json "$RESULTS_DIR/create.bin" > "$RESULTS_DIR/create_report.json"
  vegeta plot "$RESULTS_DIR/create.bin" > "$RESULTS_DIR/create_plot.html" 2>/dev/null || true
}

cmd_read() {
  mkdir -p "$RESULTS_DIR"
  wait_for_service
  [ -f perf_test/read_targets.txt ] || cmd_gen

  echo ">> read attack ${READ_RATE}/s в течение ${READ_DURATION} ..."
  vegeta attack \
    -targets=perf_test/read_targets.txt \
    -rate="$READ_RATE" -duration="$READ_DURATION" -timeout=10s \
    > "$RESULTS_DIR/read.bin"
  vegeta report "$RESULTS_DIR/read.bin" | tee "$RESULTS_DIR/read_report.txt"
  vegeta report -type=json "$RESULTS_DIR/read.bin" > "$RESULTS_DIR/read_report.json"
  vegeta plot "$RESULTS_DIR/read.bin" > "$RESULTS_DIR/read_plot.html" 2>/dev/null || true
}

cmd_stats() {
  echo ">> кол-во постов: $(post_count)"
  echo ">> размеры таблиц content_*:"
  psql_content "
    SELECT relname AS table,
           pg_size_pretty(pg_total_relation_size(relid)) AS total
    FROM pg_catalog.pg_statio_user_tables
    WHERE relname LIKE 'content_%'
    ORDER BY pg_total_relation_size(relid) DESC;"
}

cmd_reset() {
  echo ">> TRUNCATE content_post CASCADE ..."
  psql_content "TRUNCATE content_post RESTART IDENTITY CASCADE;"
  echo ">> постов в базе: $(post_count)"
}

cmd_down() { compose down; }

cmd_all() { cmd_up; cmd_gen; cmd_create; cmd_read; cmd_stats; }

case "${1:-all}" in
  up) cmd_up ;;
  gen) cmd_gen ;;
  create) cmd_create ;;
  read) cmd_read ;;
  stats) cmd_stats ;;
  reset) cmd_reset ;;
  down) cmd_down ;;
  all) cmd_all ;;
  *) echo "usage: $0 {up|gen|create|read|stats|reset|down|all}" >&2; exit 1 ;;
esac
