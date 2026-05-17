# ДЗ 3. Администрирование СУБД

## Где лежат изменения

- `docker/postgres/conf/conf.d/sporttech.conf` - PostgreSQL-конфиг, подключаемый через `docker/postgres/conf/postgresql.conf`.
- `services/auth/migrations/security/001-service-account-and-grants.sql` - сервисная роль и права auth DB.
- `services/profile/migrations/security/001-service-account-and-grants.sql` - сервисная роль и права profile DB.
- `services/content/migrations/security/001-service-account-and-grants.sql` - сервисная роль и права content DB.
- `docker-compose.yml` - запускает schema migrations, затем service-specific security migrations, затем runtime-сервисы.

Security migrations лежат рядом со schema migrations конкретного сервиса. Если потом разнести `auth`, `profile`, `content` по разным PostgreSQL-инстансам, каждый сервис уже несет свой файл прав вместе со своей схемой.

## Сервисные учетные записи

`DB_USER/DB_PASSWORD` - административный пользователь PostgreSQL, которого создает официальный Docker entrypoint через `POSTGRES_USER`. Он используется для создания баз, миграций и выдачи прав.

Runtime-сервисы используют отдельные роли:

- `AUTH_DB_APP_USER/AUTH_DB_APP_PASSWORD`
- `PROFILE_DB_APP_USER/PROFILE_DB_APP_PASSWORD`
- `CONTENT_DB_APP_USER/CONTENT_DB_APP_PASSWORD`

Минимальные права:

| Роль | База | Права |
| --- | --- | --- |
| auth app | `sporttech_auth` | `auth_user`: `SELECT, INSERT`; `auth_session`: `SELECT, INSERT, UPDATE`; sequence `auth_user_user_id_seq`: `USAGE, SELECT` |
| profile app | `sporttech_profile` | `profile`: `SELECT, INSERT, UPDATE`; `trainer_profile`: `SELECT, INSERT, UPDATE`; `trainer_sport`: `SELECT, INSERT, DELETE`; `sport_type`: `SELECT` |
| content app | `sporttech_content` | CRUD только на таблицах, которые сервис меняет; `SELECT, INSERT` на комментариях и донатах; sequence privileges только для identity-таблиц |

`CONNECT/TEMPORARY` на сервисные базы и `CREATE` в схеме `public` отозваны у `PUBLIC`, роли не имеют `SUPERUSER`, `CREATEDB`, `CREATEROLE`, `REPLICATION`.

## Пул соединений и max_connections

В Go `*sql.DB` является connection pool. Для каждого DB-сервиса настроено:

- `db_max_open_conns: 12`
- `db_max_idle_conns: 6`
- `db_conn_max_lifetime: 30m`
- `db_connect_timeout_seconds: 5`

В PostgreSQL выставлено `max_connections = 60`.

Расчет: 3 runtime-сервиса x 12 = 36 соединений. Остается около 24 соединений на миграции, healthcheck, ручной `psql`, мониторинг и административный резерв. `superuser_reserved_connections = 3` оставляет аварийный доступ администратору.

## Таймауты

- `db_statement_timeout = 5s` - API-запросы не должны выполняться минуту; это снижает эффект тяжелых запросов и DoS.
- `db_lock_timeout = 1s` - быстро падаем при долгом ожидании lock.
- `db_idle_in_transaction_session_timeout = 10s` - защита от забытых транзакций.

## SQL injection

- SQL написан вручную, без SQL builder.
- Пользовательские значения передаются через `$1..$n` и `database/sql`.
- Динамические запросы собирают только фиксированные SQL-фрагменты и placeholders.
- Для `ILIKE` экранируются `\`, `%`, `_`, используется `ESCAPE '\'`.
- На usecase-уровне есть валидация ID, пагинации, enum-значений, строк, email, username, валюты, сумм и типов файлов.

## Логи, pg_stat_statements, auto_explain, pgBadger

В `sporttech.conf` включено:

- `shared_preload_libraries = 'pg_stat_statements,auto_explain'`
- `compute_query_id = on`
- `track_io_timing = on`
- `logging_collector = on`
- `log_min_duration_statement = '500ms'`
- `auto_explain.log_min_duration = '500ms'`
- `log_line_prefix = '%t [%p]: user=%u,db=%d,app=%a,client=%h '`
- ротация логов: 1 день или 100 MB

`500ms` выбран как порог медленного запроса: для текущих CRUD/search сценариев это уже подозрительно, но быстрые запросы не будут засорять лог.

Пример разбора:

```bash
pgbadger -f stderr --prefix '%t [%p]: user=%u,db=%d,app=%a,client=%h ' /path/to/postgresql-*.log -o pgbadger.html
```

## Мониторинг

В compose есть:

- Prometheus: `http://localhost:8090`
- Grafana: `http://localhost:8100`
- `node-exporter` для host CPU/RAM.
- `cadvisor` для CPU/RAM/network контейнеров.
- `postgres-exporter` для PostgreSQL metrics.
- `/metrics` на Go-сервисах для RPS/latency HTTP и gRPC.

Prometheus scrape targets лежат в `docker/prometheus/prometheus.yml`.
