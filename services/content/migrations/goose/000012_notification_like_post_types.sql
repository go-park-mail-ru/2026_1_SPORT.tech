-- +goose Up
ALTER TABLE content_notification
  DROP CONSTRAINT IF EXISTS content_notification_type_check;

ALTER TABLE content_notification
  ADD CONSTRAINT content_notification_type_check
  CHECK (type IN ('comment', 'donation', 'like', 'post', 'subscription'));

-- +goose Down
ALTER TABLE content_notification
  DROP CONSTRAINT IF EXISTS content_notification_type_check;

ALTER TABLE content_notification
  ADD CONSTRAINT content_notification_type_check
  CHECK (type IN ('comment', 'donation', 'subscription'));
