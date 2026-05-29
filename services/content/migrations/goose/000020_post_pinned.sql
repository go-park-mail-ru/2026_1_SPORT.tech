-- +goose Up
ALTER TABLE content_post
ADD COLUMN is_pinned BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX content_post_one_pinned_per_author
ON content_post(author_user_id)
WHERE is_pinned;

-- +goose Down
DROP INDEX IF EXISTS content_post_one_pinned_per_author;

ALTER TABLE content_post
DROP COLUMN IF EXISTS is_pinned;
