package postgres

import (
	"context"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"time"
)

func (repository *Repository) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	now := time.Now().UTC()

	const query = `
		INSERT INTO content_comment (
			post_id,
			author_user_id,
			body,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $4)
		RETURNING comment_id, created_at, updated_at
	`

	created := comment
	err := repository.db.QueryRowContext(
		ctx,
		query,
		comment.PostID,
		comment.AuthorUserID,
		comment.Body,
		now,
	).Scan(&created.CommentID, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return domain.Comment{}, err
	}

	return created, nil
}

func (repository *Repository) ListComments(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
	const query = `
		SELECT
			comment_id,
			post_id,
			author_user_id,
			body,
			created_at,
			updated_at
		FROM content_comment
		WHERE post_id = $1
		ORDER BY created_at ASC, comment_id ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := repository.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var comment domain.Comment
		if err := rows.Scan(
			&comment.CommentID,
			&comment.PostID,
			&comment.AuthorUserID,
			&comment.Body,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}
