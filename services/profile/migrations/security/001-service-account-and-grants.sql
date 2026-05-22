\set ON_ERROR_STOP on

SELECT format(
  'CREATE ROLE %I WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION',
  :'profile_app_user',
  :'profile_app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'profile_app_user') \gexec

ALTER ROLE :"profile_app_user" WITH LOGIN PASSWORD :'profile_app_password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION;

REVOKE CONNECT, TEMPORARY ON DATABASE sporttech_profile FROM PUBLIC;
GRANT CONNECT ON DATABASE sporttech_profile TO :"profile_app_user";

\connect sporttech_profile

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO :"profile_app_user";

GRANT SELECT, INSERT, UPDATE ON TABLE profile TO :"profile_app_user";
GRANT SELECT, INSERT, UPDATE ON TABLE trainer_profile TO :"profile_app_user";
GRANT SELECT, INSERT, DELETE ON TABLE trainer_sport TO :"profile_app_user";
GRANT SELECT ON TABLE sport_type TO :"profile_app_user";

ALTER ROLE :"profile_app_user" IN DATABASE sporttech_profile SET statement_timeout = '5s';
ALTER ROLE :"profile_app_user" IN DATABASE sporttech_profile SET lock_timeout = '1s';
ALTER ROLE :"profile_app_user" IN DATABASE sporttech_profile SET idle_in_transaction_session_timeout = '10s';
ALTER ROLE :"profile_app_user" IN DATABASE sporttech_profile SET search_path = public;
