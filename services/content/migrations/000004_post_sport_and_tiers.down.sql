DROP INDEX IF EXISTS content_post_sport_type_id_idx;

ALTER TABLE content_post
DROP CONSTRAINT IF EXISTS content_post_required_subscription_tier_fkey;

DROP TABLE IF EXISTS content_subscription_tier;

ALTER TABLE content_post
DROP COLUMN IF EXISTS sport_type_id;
