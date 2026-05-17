\set ON_ERROR_STOP on

SELECT format(
  'CREATE ROLE %I WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION',
  :'auth_app_user',
  :'auth_app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'auth_app_user') \gexec

ALTER ROLE :"auth_app_user" WITH LOGIN PASSWORD :'auth_app_password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION;

REVOKE CONNECT, TEMPORARY ON DATABASE sporttech_auth FROM PUBLIC;
GRANT CONNECT ON DATABASE sporttech_auth TO :"auth_app_user";

\connect sporttech_auth

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO :"auth_app_user";

GRANT SELECT, INSERT ON TABLE auth_user TO :"auth_app_user";
GRANT SELECT, INSERT, UPDATE ON TABLE auth_session TO :"auth_app_user";
GRANT USAGE, SELECT ON SEQUENCE auth_user_user_id_seq TO :"auth_app_user";

ALTER ROLE :"auth_app_user" IN DATABASE sporttech_auth SET statement_timeout = '5s';
ALTER ROLE :"auth_app_user" IN DATABASE sporttech_auth SET lock_timeout = '1s';
ALTER ROLE :"auth_app_user" IN DATABASE sporttech_auth SET idle_in_transaction_session_timeout = '10s';
ALTER ROLE :"auth_app_user" IN DATABASE sporttech_auth SET search_path = public;
