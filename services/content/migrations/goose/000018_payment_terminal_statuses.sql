-- +goose Up
ALTER TABLE content_payment
  DROP CONSTRAINT IF EXISTS content_payment_status_check;

ALTER TABLE content_payment
  ADD CONSTRAINT content_payment_status_check
  CHECK (status IN ('pending', 'confirmed', 'canceled', 'failed', 'expired'));

-- +goose Down
UPDATE content_payment
  SET status = 'pending'
  WHERE status IN ('canceled', 'failed', 'expired');

ALTER TABLE content_payment
  DROP CONSTRAINT IF EXISTS content_payment_status_check;

ALTER TABLE content_payment
  ADD CONSTRAINT content_payment_status_check
  CHECK (status IN ('pending', 'confirmed'));
