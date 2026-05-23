-- +goose Up

-- Add chat_enabled flag to subscription tiers
ALTER TABLE content_subscription_tier
ADD COLUMN chat_enabled BOOLEAN NOT NULL DEFAULT false;

-- Chat messages table
CREATE TABLE content_chat_message (
  message_id      BIGSERIAL PRIMARY KEY,
  sender_user_id  BIGINT NOT NULL,
  receiver_user_id BIGINT NOT NULL,
  body            TEXT NOT NULL CHECK (char_length(body) BETWEEN 1 AND 4000),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  is_read         BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX content_chat_message_conversation_idx
  ON content_chat_message (
    LEAST(sender_user_id, receiver_user_id),
    GREATEST(sender_user_id, receiver_user_id),
    created_at DESC
  );

CREATE INDEX content_chat_message_receiver_unread_idx
  ON content_chat_message (receiver_user_id, is_read)
  WHERE is_read = false;

-- +goose Down
DROP INDEX IF EXISTS content_chat_message_receiver_unread_idx;
DROP INDEX IF EXISTS content_chat_message_conversation_idx;
DROP TABLE IF EXISTS content_chat_message;

ALTER TABLE content_subscription_tier
DROP COLUMN IF EXISTS chat_enabled;
