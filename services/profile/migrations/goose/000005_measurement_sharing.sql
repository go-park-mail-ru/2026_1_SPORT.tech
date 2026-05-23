-- +goose Up
-- Таблица доступа тренеров к замерам клиента.
-- Клиент явно разрешает конкретным тренерам просматривать свои замеры.
CREATE TABLE measurement_sharing (
    client_user_id  BIGINT NOT NULL,
    trainer_user_id BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (client_user_id, trainer_user_id)
);

CREATE INDEX measurement_sharing_client_idx ON measurement_sharing(client_user_id);

-- +goose Down
DROP INDEX IF EXISTS measurement_sharing_client_idx;
DROP TABLE IF EXISTS measurement_sharing;
