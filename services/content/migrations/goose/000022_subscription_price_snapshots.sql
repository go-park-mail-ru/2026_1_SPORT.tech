-- +goose Up
ALTER TABLE content_subscription
  ADD COLUMN IF NOT EXISTS tier_name_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS price_snapshot INTEGER,
  ADD COLUMN IF NOT EXISTS price_change_requires_resubscribe BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE content_subscription subscription
SET tier_name_snapshot = COALESCE(subscription.tier_name_snapshot, tier.name),
    price_snapshot = COALESCE(subscription.price_snapshot, tier.price)
FROM content_subscription_tier tier
WHERE tier.trainer_user_id = subscription.trainer_user_id
  AND tier.tier_id = subscription.tier_id
  AND (
    subscription.tier_name_snapshot IS NULL
    OR subscription.price_snapshot IS NULL
  );

UPDATE content_subscription
SET tier_name_snapshot = COALESCE(tier_name_snapshot, ''),
    price_snapshot = COALESCE(price_snapshot, 0);

ALTER TABLE content_subscription
  ALTER COLUMN tier_name_snapshot SET NOT NULL,
  ALTER COLUMN price_snapshot SET NOT NULL,
  ADD CONSTRAINT content_subscription_price_snapshot_non_negative CHECK (price_snapshot >= 0);

-- +goose Down
ALTER TABLE content_subscription
  DROP CONSTRAINT IF EXISTS content_subscription_price_snapshot_non_negative,
  DROP COLUMN IF EXISTS price_change_requires_resubscribe,
  DROP COLUMN IF EXISTS price_snapshot,
  DROP COLUMN IF EXISTS tier_name_snapshot;
