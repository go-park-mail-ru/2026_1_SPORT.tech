-- +goose Up
ALTER TABLE content_payment
  ADD COLUMN IF NOT EXISTS tier_id BIGINT,
  ADD COLUMN IF NOT EXISTS subscription_id BIGINT REFERENCES content_subscription(subscription_id);

-- +goose Down
ALTER TABLE content_payment
  DROP COLUMN IF EXISTS subscription_id,
  DROP COLUMN IF EXISTS tier_id;
