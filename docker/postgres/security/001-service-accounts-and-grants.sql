\set ON_ERROR_STOP on

SELECT format(
  'CREATE ROLE %I WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION',
  :'auth_app_user',
  :'auth_app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'auth_app_user') \gexec

SELECT format(
  'CREATE ROLE %I WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION',
  :'profile_app_user',
  :'profile_app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'profile_app_user') \gexec

SELECT format(
  'CREATE ROLE %I WITH LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION',
  :'content_app_user',
  :'content_app_password'
)
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'content_app_user') \gexec

ALTER ROLE :"auth_app_user" WITH LOGIN PASSWORD :'auth_app_password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION;
ALTER ROLE :"profile_app_user" WITH LOGIN PASSWORD :'profile_app_password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION;
ALTER ROLE :"content_app_user" WITH LOGIN PASSWORD :'content_app_password' NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION;

REVOKE CONNECT, TEMPORARY ON DATABASE sporttech_auth FROM PUBLIC;
REVOKE CONNECT, TEMPORARY ON DATABASE sporttech_profile FROM PUBLIC;
REVOKE CONNECT, TEMPORARY ON DATABASE sporttech_content FROM PUBLIC;
GRANT CONNECT ON DATABASE sporttech_auth TO :"auth_app_user";
GRANT CONNECT ON DATABASE sporttech_profile TO :"profile_app_user";
GRANT CONNECT ON DATABASE sporttech_content TO :"content_app_user";

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

\connect sporttech_content

CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

REVOKE CREATE ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO :"content_app_user";
GRANT USAGE ON TYPE content_block_kind TO :"content_app_user";

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE content_post TO :"content_app_user";
GRANT SELECT, INSERT, DELETE ON TABLE content_post_block TO :"content_app_user";
GRANT SELECT, INSERT ON TABLE content_comment TO :"content_app_user";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE content_post_like TO :"content_app_user";
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE content_subscription_tier TO :"content_app_user";
GRANT SELECT, INSERT, UPDATE ON TABLE content_subscription TO :"content_app_user";
GRANT SELECT, INSERT ON TABLE content_donation TO :"content_app_user";

GRANT USAGE, SELECT ON SEQUENCE content_post_post_id_seq TO :"content_app_user";
GRANT USAGE, SELECT ON SEQUENCE content_post_block_post_block_id_seq TO :"content_app_user";
GRANT USAGE, SELECT ON SEQUENCE content_comment_comment_id_seq TO :"content_app_user";
GRANT USAGE, SELECT ON SEQUENCE content_subscription_subscription_id_seq TO :"content_app_user";
GRANT USAGE, SELECT ON SEQUENCE content_donation_donation_id_seq TO :"content_app_user";

ALTER ROLE :"content_app_user" IN DATABASE sporttech_content SET statement_timeout = '5s';
ALTER ROLE :"content_app_user" IN DATABASE sporttech_content SET lock_timeout = '1s';
ALTER ROLE :"content_app_user" IN DATABASE sporttech_content SET idle_in_transaction_session_timeout = '10s';
ALTER ROLE :"content_app_user" IN DATABASE sporttech_content SET search_path = public;
