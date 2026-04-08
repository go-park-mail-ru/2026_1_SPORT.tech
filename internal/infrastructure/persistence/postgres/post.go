package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
	"github.com/lib/pq"
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

func (repository *PostRepository) Create(ctx context.Context, trainerID int64, command usecase.CreatePostCommand) (int64, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	const createPostQuery = `
		INSERT INTO post (trainer_id, min_tier_id, title, text_content)
		VALUES ($1, $2, $3, $4)
		RETURNING post_id
	`

	var postID int64
	if err := queryRowContext(
		ctx,
		tx,
		repository.logger,
		"post.create",
		createPostQuery,
		trainerID,
		command.MinTierID,
		command.Title,
		command.TextContent,
	).Scan(&postID); err != nil {
		return 0, mapPostError(err)
	}

	const createAttachmentQuery = `
		INSERT INTO post_attachment (post_id, kind, file_url)
		VALUES ($1, $2, $3)
	`

	for _, attachment := range command.Attachments {
		if _, err := execContext(
			ctx,
			tx,
			repository.logger,
			"post.create_attachment",
			createAttachmentQuery,
			postID,
			attachment.Kind,
			attachment.FileURL,
		); err != nil {
			return 0, mapPostError(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return postID, nil
}

func (repository *PostRepository) Update(ctx context.Context, trainerID int64, postID int64, command usecase.UpdatePostCommand) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const getPostForUpdateQuery = `
		SELECT trainer_id, min_tier_id, title, text_content
		FROM post
		WHERE post_id = $1
		FOR UPDATE
	`

	var (
		postTrainerID      int64
		currentMinTierID   sql.NullInt64
		currentTitle       string
		currentTextContent string
	)

	err = queryRowContext(
		ctx,
		tx,
		repository.logger,
		"post.get_for_update",
		getPostForUpdateQuery,
		postID,
	).Scan(
		&postTrainerID,
		&currentMinTierID,
		&currentTitle,
		&currentTextContent,
	)
	if err != nil {
		return err
	}

	if postTrainerID != trainerID {
		return usecase.ErrPostForbidden
	}

	var updatedMinTierID *int64
	if currentMinTierID.Valid {
		updatedMinTierID = &currentMinTierID.Int64
	}
	if command.HasMinTierID {
		updatedMinTierID = command.MinTierID
	}

	updatedTitle := currentTitle
	if command.Title != nil {
		updatedTitle = *command.Title
	}

	updatedTextContent := currentTextContent
	if command.TextContent != nil {
		updatedTextContent = *command.TextContent
	}

	const updatePostQuery = `
		UPDATE post
		SET min_tier_id = $3,
		    title = $4,
		    text_content = $5,
		    updated_at = now()
		WHERE post_id = $1
		  AND trainer_id = $2
	`

	if _, err := execContext(
		ctx,
		tx,
		repository.logger,
		"post.update",
		updatePostQuery,
		postID,
		trainerID,
		updatedMinTierID,
		updatedTitle,
		updatedTextContent,
	); err != nil {
		return mapPostError(err)
	}

	if command.HasAttachments {
		const deleteAttachmentsQuery = `
			DELETE FROM post_attachment
			WHERE post_id = $1
		`

		if _, err := execContext(
			ctx,
			tx,
			repository.logger,
			"post.delete_attachments",
			deleteAttachmentsQuery,
			postID,
		); err != nil {
			return err
		}

		const createAttachmentQuery = `
			INSERT INTO post_attachment (post_id, kind, file_url)
			VALUES ($1, $2, $3)
		`

		for _, attachment := range command.Attachments {
			if _, err := execContext(
				ctx,
				tx,
				repository.logger,
				"post.create_attachment",
				createAttachmentQuery,
				postID,
				attachment.Kind,
				attachment.FileURL,
			); err != nil {
				return mapPostError(err)
			}
		}
	}

	return tx.Commit()
}

func (repository *PostRepository) Delete(ctx context.Context, trainerID int64, postID int64) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const getPostForDeleteQuery = `
		SELECT trainer_id
		FROM post
		WHERE post_id = $1
		FOR UPDATE
	`

	var postTrainerID int64
	err = queryRowContext(
		ctx,
		tx,
		repository.logger,
		"post.get_for_delete",
		getPostForDeleteQuery,
		postID,
	).Scan(&postTrainerID)
	if err != nil {
		return err
	}

	if postTrainerID != trainerID {
		return usecase.ErrPostForbidden
	}

	const deletePostQuery = `
		DELETE FROM post
		WHERE post_id = $1
		  AND trainer_id = $2
	`

	if _, err := execContext(
		ctx,
		tx,
		repository.logger,
		"post.delete",
		deletePostQuery,
		postID,
		trainerID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func mapPostError(err error) error {
	var pqError *pq.Error
	if !errors.As(err, &pqError) {
		return err
	}

	switch {
	case pqError.Code == "23503" && pqError.Constraint == "post_min_tier_fk":
		return usecase.ErrPostTierNotFound
	default:
		return err
	}
}
