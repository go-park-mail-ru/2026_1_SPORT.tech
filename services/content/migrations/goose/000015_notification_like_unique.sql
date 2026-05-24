-- +goose Up
-- Предотвращаем дублирование уведомлений о лайке:
-- один актор может создать не более одного уведомления типа 'like' для конкретного поста и получателя.
CREATE UNIQUE INDEX content_notification_like_unique_idx
    ON content_notification (user_id, actor_user_id, post_id)
    WHERE type = 'like' AND post_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS content_notification_like_unique_idx;
