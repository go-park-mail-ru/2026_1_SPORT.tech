-- +goose Up
ALTER TABLE content_subscription
  ADD COLUMN stripe_subscription_id TEXT,
  ADD COLUMN current_period_end TIMESTAMPTZ,
  ADD COLUMN auto_renew BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX content_subscription_stripe_subscription_id_key
  ON content_subscription(stripe_subscription_id)
  WHERE stripe_subscription_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS content_subscription_stripe_subscription_id_key;

ALTER TABLE content_subscription
  DROP COLUMN IF EXISTS auto_renew,
  DROP COLUMN IF EXISTS current_period_end,
  DROP COLUMN IF EXISTS stripe_subscription_id;
