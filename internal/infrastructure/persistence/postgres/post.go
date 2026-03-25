package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

type PostRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPostRepository(db *sql.DB, logger *slog.Logger) *PostRepository {
	return &PostRepository{
		db:     db,
		logger: logger,
	}
}

func (repository *PostRepository) ListProfilePosts(ctx context.Context, profileUserID int64, currentUserID int64) ([]domain.PostListItem, error) {
	const query = `
		SELECT
			p.post_id,
			p.trainer_id,
			p.min_tier_id,
			p.title,
			p.created_at,
			CASE
				WHEN p.min_tier_id IS NULL THEN true
				WHEN p.trainer_id = $2 THEN true
				WHEN EXISTS (
					SELECT 1
					FROM user_subscription us
					JOIN subscription_tier viewer_tier
					  ON viewer_tier.subscription_tier_id = us.subscription_tier_id
					JOIN subscription_tier post_tier
					  ON post_tier.subscription_tier_id = p.min_tier_id
					WHERE us.subscriber_user_id = $2
					  AND us.expires_at > now()
					  AND viewer_tier.trainer_id = p.trainer_id
					  AND post_tier.trainer_id = p.trainer_id
					  AND viewer_tier.level_rank >= post_tier.level_rank
				) THEN true
				ELSE false
			END AS can_view
		FROM post p
		WHERE p.trainer_id = $1
		ORDER BY p.created_at DESC, p.post_id DESC
	`

	rows, err := queryContext(ctx, repository.db, repository.logger, "post.list_profile_posts", query, profileUserID, currentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]domain.PostListItem, 0)
	for rows.Next() {
		var (
			post      domain.PostListItem
			minTierID sql.NullInt64
		)

		if err := rows.Scan(
			&post.PostID,
			&post.TrainerID,
			&minTierID,
			&post.Title,
			&post.CreatedAt,
			&post.CanView,
		); err != nil {
			return nil, err
		}

		if minTierID.Valid {
			post.MinTierID = &minTierID.Int64
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (repository *PostRepository) GetByID(ctx context.Context, postID int64, currentUserID int64) (domain.Post, error) {
	const postQuery = `
		SELECT
			p.post_id,
			p.trainer_id,
			p.min_tier_id,
			p.title,
			p.text_content,
			p.created_at,
			p.updated_at,
			CASE
				WHEN p.min_tier_id IS NULL THEN true
				WHEN p.trainer_id = $2 THEN true
				WHEN EXISTS (
					SELECT 1
					FROM user_subscription us
					JOIN subscription_tier viewer_tier
					  ON viewer_tier.subscription_tier_id = us.subscription_tier_id
					JOIN subscription_tier post_tier
					  ON post_tier.subscription_tier_id = p.min_tier_id
					WHERE us.subscriber_user_id = $2
					  AND us.expires_at > now()
					  AND viewer_tier.trainer_id = p.trainer_id
					  AND post_tier.trainer_id = p.trainer_id
					  AND viewer_tier.level_rank >= post_tier.level_rank
				) THEN true
				ELSE false
			END AS can_view
		FROM post p
		WHERE p.post_id = $1
	`

	var (
		post      domain.Post
		minTierID sql.NullInt64
	)

	err := queryRowContext(ctx, repository.db, repository.logger, "post.get_by_id", postQuery, postID, currentUserID).Scan(
		&post.PostID,
		&post.TrainerID,
		&minTierID,
		&post.Title,
		&post.TextContent,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.CanView,
	)
	if err != nil {
		return domain.Post{}, err
	}

	if minTierID.Valid {
		post.MinTierID = &minTierID.Int64
	}

	const attachmentQuery = `
		SELECT post_attachment_id, kind, file_url
		FROM post_attachment
		WHERE post_id = $1
		ORDER BY post_attachment_id
	`

	rows, err := queryContext(ctx, repository.db, repository.logger, "post.list_attachments", attachmentQuery, postID)
	if err != nil {
		return domain.Post{}, err
	}
	defer rows.Close()

	post.Attachments = make([]domain.PostAttachment, 0)
	for rows.Next() {
		var attachment domain.PostAttachment
		if err := rows.Scan(
			&attachment.PostAttachmentID,
			&attachment.Kind,
			&attachment.FileURL,
		); err != nil {
			return domain.Post{}, err
		}

		post.Attachments = append(post.Attachments, attachment)
	}

	if err := rows.Err(); err != nil {
		return domain.Post{}, err
	}

	return post, nil
}
