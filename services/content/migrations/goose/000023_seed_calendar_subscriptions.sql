-- +goose Up
UPDATE content_subscription_tier
SET calendar_enabled = TRUE,
    updated_at = NOW()
WHERE (trainer_user_id, tier_id) IN (
  (1001, 2),
  (1003, 1),
  (1004, 2),
  (1005, 3),
  (1006, 2),
  (1007, 2)
);

WITH desired_subscriptions (
  subscription_id,
  client_user_id,
  trainer_user_id,
  tier_id,
  tier_name_snapshot,
  price_snapshot
) AS (
  VALUES
    (2401, 1002, 1001, 2, 'Продвинутый', 1500),
    (2404, 1002, 1003, 1, 'Базовый', 600),
    (2405, 1002, 1004, 2, 'Глубокая практика', 1200),
    (2406, 1002, 1005, 3, 'Премиум', 2200),
    (2407, 1002, 1006, 2, 'Прогрессия', 1300),
    (2408, 1002, 1007, 2, 'План подготовки', 1500)
),
updated_subscriptions AS (
  UPDATE content_subscription subscription
  SET tier_id = desired.tier_id,
      active = TRUE,
      expires_at = NOW() + INTERVAL '30 days',
      auto_renew = FALSE,
      tier_name_snapshot = desired.tier_name_snapshot,
      price_snapshot = desired.price_snapshot,
      price_change_requires_resubscribe = FALSE,
      updated_at = NOW()
  FROM desired_subscriptions desired
  WHERE subscription.client_user_id = desired.client_user_id
    AND subscription.trainer_user_id = desired.trainer_user_id
    AND subscription.active = TRUE
  RETURNING desired.subscription_id
)
INSERT INTO content_subscription (
  subscription_id,
  client_user_id,
  trainer_user_id,
  tier_id,
  active,
  expires_at,
  auto_renew,
  tier_name_snapshot,
  price_snapshot
)
SELECT desired.subscription_id,
       desired.client_user_id,
       desired.trainer_user_id,
       desired.tier_id,
       TRUE,
       NOW() + INTERVAL '30 days',
       FALSE,
       desired.tier_name_snapshot,
       desired.price_snapshot
FROM desired_subscriptions desired
WHERE NOT EXISTS (
  SELECT 1
  FROM updated_subscriptions updated
  WHERE updated.subscription_id = desired.subscription_id
)
ON CONFLICT (subscription_id) DO UPDATE
SET client_user_id = EXCLUDED.client_user_id,
    trainer_user_id = EXCLUDED.trainer_user_id,
    tier_id = EXCLUDED.tier_id,
    active = EXCLUDED.active,
    expires_at = EXCLUDED.expires_at,
    auto_renew = EXCLUDED.auto_renew,
    tier_name_snapshot = EXCLUDED.tier_name_snapshot,
    price_snapshot = EXCLUDED.price_snapshot,
    price_change_requires_resubscribe = FALSE,
    updated_at = NOW();

SELECT setval(
  pg_get_serial_sequence('content_subscription', 'subscription_id'),
  (SELECT GREATEST(COALESCE(MAX(subscription_id), 1), 2408) FROM content_subscription),
  TRUE
);

INSERT INTO content_meeting_slot (slot_id, trainer_user_id, starts_at)
VALUES
  (2701, 1001, date_trunc('day', NOW()) + INTERVAL '2 days 10 hours'),
  (2702, 1001, date_trunc('day', NOW()) + INTERVAL '4 days 18 hours'),
  (2703, 1001, date_trunc('day', NOW()) + INTERVAL '7 days 10 hours'),
  (2704, 1003, date_trunc('day', NOW()) + INTERVAL '2 days 11 hours'),
  (2705, 1003, date_trunc('day', NOW()) + INTERVAL '5 days 17 hours'),
  (2706, 1003, date_trunc('day', NOW()) + INTERVAL '8 days 11 hours'),
  (2707, 1004, date_trunc('day', NOW()) + INTERVAL '3 days 9 hours'),
  (2708, 1004, date_trunc('day', NOW()) + INTERVAL '5 days 16 hours'),
  (2709, 1004, date_trunc('day', NOW()) + INTERVAL '9 days 9 hours'),
  (2710, 1005, date_trunc('day', NOW()) + INTERVAL '3 days 12 hours'),
  (2711, 1005, date_trunc('day', NOW()) + INTERVAL '6 days 19 hours'),
  (2712, 1005, date_trunc('day', NOW()) + INTERVAL '10 days 12 hours'),
  (2713, 1006, date_trunc('day', NOW()) + INTERVAL '4 days 10 hours'),
  (2714, 1006, date_trunc('day', NOW()) + INTERVAL '6 days 18 hours'),
  (2715, 1006, date_trunc('day', NOW()) + INTERVAL '11 days 10 hours'),
  (2716, 1007, date_trunc('day', NOW()) + INTERVAL '4 days 11 hours'),
  (2717, 1007, date_trunc('day', NOW()) + INTERVAL '7 days 17 hours'),
  (2718, 1007, date_trunc('day', NOW()) + INTERVAL '12 days 11 hours')
ON CONFLICT (slot_id) DO UPDATE
SET trainer_user_id = EXCLUDED.trainer_user_id,
    starts_at = EXCLUDED.starts_at;

SELECT setval(
  pg_get_serial_sequence('content_meeting_slot', 'slot_id'),
  (SELECT GREATEST(COALESCE(MAX(slot_id), 1), 2718) FROM content_meeting_slot),
  TRUE
);

-- +goose Down
DELETE FROM content_meeting_slot
WHERE slot_id BETWEEN 2701 AND 2718;

DELETE FROM content_subscription
WHERE subscription_id BETWEEN 2404 AND 2408;

UPDATE content_subscription_tier
SET calendar_enabled = FALSE,
    updated_at = NOW()
WHERE (trainer_user_id, tier_id) IN (
  (1001, 2),
  (1003, 1),
  (1004, 2),
  (1005, 3),
  (1006, 2),
  (1007, 2)
);
