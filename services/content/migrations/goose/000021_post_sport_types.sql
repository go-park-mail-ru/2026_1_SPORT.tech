-- +goose Up
CREATE TABLE content_post_sport_type (
  post_id BIGINT NOT NULL REFERENCES content_post(post_id) ON DELETE CASCADE,
  sport_type_id BIGINT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (post_id, sport_type_id)
);

INSERT INTO content_post_sport_type (post_id, sport_type_id, created_at)
SELECT post_id, sport_type_id, created_at
FROM content_post
WHERE sport_type_id IS NOT NULL
ON CONFLICT (post_id, sport_type_id) DO NOTHING;

CREATE INDEX content_post_sport_type_sport_idx
ON content_post_sport_type(sport_type_id, post_id);

-- +goose Down
DROP INDEX IF EXISTS content_post_sport_type_sport_idx;
DROP TABLE IF EXISTS content_post_sport_type;
