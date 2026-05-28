-- +goose Up
CREATE TABLE notification_preferences (
    user_id       BIGINT PRIMARY KEY,
    comments      BOOLEAN NOT NULL DEFAULT true,
    likes         BOOLEAN NOT NULL DEFAULT true,
    donations     BOOLEAN NOT NULL DEFAULT true,
    posts         BOOLEAN NOT NULL DEFAULT true,
    subscriptions BOOLEAN NOT NULL DEFAULT true,
    meetings      BOOLEAN NOT NULL DEFAULT true,
    email_digest  BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS notification_preferences;
