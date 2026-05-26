-- +goose Up
ALTER TABLE content_payment
  DROP CONSTRAINT IF EXISTS content_payment_subscription_id_key;

-- +goose Down
ALTER TABLE content_payment
  ADD CONSTRAINT content_payment_subscription_id_key UNIQUE (subscription_id);
