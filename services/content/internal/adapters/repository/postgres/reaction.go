package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"time"
)

func (repository *Repository) UpsertLike(ctx context.Context, postID int64, userID int64) (bool, error) {
	now := time.Now().UTC()

	var insertedPostID int64
	err := repository.db.QueryRowContext(
		ctx,
		`
			INSERT INTO content_post_like (post_id, user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $3)
			ON CONFLICT (post_id, user_id)
			DO NOTHING
			RETURNING post_id
		`,
		postID,
		userID,
		now,
	).Scan(&insertedPostID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (repository *Repository) DeleteLike(ctx context.Context, postID int64, userID int64) error {
	_, err := repository.db.ExecContext(
		ctx,
		`DELETE FROM content_post_like WHERE post_id = $1 AND user_id = $2`,
		postID,
		userID,
	)

	return err
}

func (repository *Repository) ListPostLikes(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.PostLike, error) {
	const query = `
		SELECT
			post_id,
			user_id,
			created_at
		FROM content_post_like
		WHERE post_id = $1
		ORDER BY created_at DESC, user_id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := repository.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	likes := make([]domain.PostLike, 0)
	for rows.Next() {
		var like domain.PostLike
		if err := rows.Scan(&like.PostID, &like.UserID, &like.CreatedAt); err != nil {
			return nil, err
		}
		likes = append(likes, like)
	}

	return likes, rows.Err()
}

func (repository *Repository) GetPostLikeState(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
	const query = `
		SELECT
			p.post_id,
			COALESCE(COUNT(pl.user_id), 0) AS likes_count,
			EXISTS (
				SELECT 1
				FROM content_post_like viewer_like
				WHERE viewer_like.post_id = p.post_id
					AND viewer_like.user_id = $2
			) AS is_liked
		FROM content_post p
		LEFT JOIN content_post_like pl ON pl.post_id = p.post_id
		WHERE p.post_id = $1
		GROUP BY p.post_id
	`

	var state domain.PostLikeState
	err := repository.db.QueryRowContext(ctx, query, postID, userID).Scan(
		&state.PostID,
		&state.LikesCount,
		&state.IsLiked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PostLikeState{}, domain.ErrPostNotFound
		}
		return domain.PostLikeState{}, err
	}

	return state, nil
}
