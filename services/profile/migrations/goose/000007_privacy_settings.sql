-- +goose Up
CREATE TABLE profile_privacy_settings (
    user_id                   BIGINT PRIMARY KEY,
    show_profile_in_search    BOOLEAN NOT NULL DEFAULT true,
    allow_measurement_sharing BOOLEAN NOT NULL DEFAULT true,
    show_activity_status      BOOLEAN NOT NULL DEFAULT true,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS profile_privacy_settings;
