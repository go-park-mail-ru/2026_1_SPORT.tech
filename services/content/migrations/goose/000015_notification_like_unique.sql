-- +goose Up
CREATE UNIQUE INDEX content_notification_like_unique_idx
    ON content_notification (user_id, actor_user_id, post_id)
    WHERE type = 'like' AND post_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS content_notification_like_unique_idx;
