-- +goose Up
ALTER TABLE content_payment
  DROP CONSTRAINT IF EXISTS content_payment_provider_check;

ALTER TABLE content_payment
  ADD CONSTRAINT content_payment_provider_check
  CHECK (provider IN ('stripe'));

-- +goose Down
ALTER TABLE content_payment
  DROP CONSTRAINT IF EXISTS content_payment_provider_check;

ALTER TABLE content_payment
  ADD CONSTRAINT content_payment_provider_check
  CHECK (provider IN ('mock'));
