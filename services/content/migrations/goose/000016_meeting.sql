-- +goose Up

ALTER TABLE content_subscription_tier
ADD COLUMN calendar_enabled BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE content_notification
  DROP CONSTRAINT IF EXISTS content_notification_type_check;

ALTER TABLE content_notification
  ADD CONSTRAINT content_notification_type_check
  CHECK (type IN ('comment', 'donation', 'like', 'post', 'subscription', 'meeting'));

CREATE TABLE content_meeting_availability_rule (
  rule_id         BIGSERIAL PRIMARY KEY,
  trainer_user_id BIGINT NOT NULL,
  weekday         SMALLINT NOT NULL CHECK (weekday BETWEEN 0 AND 6),
  start_hour      SMALLINT NOT NULL CHECK (start_hour BETWEEN 0 AND 23),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (trainer_user_id, weekday, start_hour)
);

CREATE TABLE content_meeting_slot (
  slot_id         BIGSERIAL PRIMARY KEY,
  trainer_user_id BIGINT NOT NULL,
  starts_at       TIMESTAMPTZ NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (trainer_user_id, starts_at)
);

CREATE INDEX content_meeting_slot_trainer_starts_idx
  ON content_meeting_slot (trainer_user_id, starts_at);

CREATE TABLE content_meeting_booking (
  booking_id           BIGSERIAL PRIMARY KEY,
  trainer_user_id      BIGINT NOT NULL,
  client_user_id       BIGINT NOT NULL,
  starts_at            TIMESTAMPTZ NOT NULL,
  ends_at              TIMESTAMPTZ NOT NULL,
  status               TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'cancelled')),
  created_by_user_id   BIGINT NOT NULL,
  note                 TEXT CHECK (note IS NULL OR char_length(note) BETWEEN 1 AND 1000),
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  cancelled_at         TIMESTAMPTZ,
  cancelled_by_user_id BIGINT,
  CHECK (ends_at > starts_at)
);

CREATE UNIQUE INDEX content_meeting_booking_trainer_start_unique_idx
  ON content_meeting_booking (trainer_user_id, starts_at)
  WHERE status = 'confirmed';

CREATE INDEX content_meeting_booking_trainer_range_idx
  ON content_meeting_booking (trainer_user_id, starts_at, ends_at)
  WHERE status = 'confirmed';

CREATE INDEX content_meeting_booking_client_idx
  ON content_meeting_booking (client_user_id, starts_at);

-- +goose Down
DROP INDEX IF EXISTS content_meeting_booking_client_idx;
DROP INDEX IF EXISTS content_meeting_booking_trainer_range_idx;
DROP INDEX IF EXISTS content_meeting_booking_trainer_start_unique_idx;
DROP TABLE IF EXISTS content_meeting_booking;
DROP INDEX IF EXISTS content_meeting_slot_trainer_starts_idx;
DROP TABLE IF EXISTS content_meeting_slot;
DROP TABLE IF EXISTS content_meeting_availability_rule;

ALTER TABLE content_notification
  DROP CONSTRAINT IF EXISTS content_notification_type_check;

ALTER TABLE content_notification
  ADD CONSTRAINT content_notification_type_check
  CHECK (type IN ('comment', 'donation', 'like', 'post', 'subscription'));

ALTER TABLE content_subscription_tier
DROP COLUMN IF EXISTS calendar_enabled;
