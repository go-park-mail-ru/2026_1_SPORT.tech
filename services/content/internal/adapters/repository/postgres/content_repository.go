package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"github.com/lib/pq"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) CreatePost(ctx context.Context, post domain.Post) (int64, error) {
	now := time.Now().UTC()
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	const insertPostQuery = `
		INSERT INTO content_post (
			author_user_id,
			title,
			required_subscription_level,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $4)
		RETURNING post_id
	`

	var postID int64
	err = tx.QueryRowContext(
		ctx,
		insertPostQuery,
		post.AuthorUserID,
		post.Title,
		nullInt32(post.RequiredSubscriptionLevel),
		now,
	).Scan(&postID)
	if err != nil {
		return 0, err
	}

	if err := insertBlocks(ctx, tx, postID, post.Blocks, now); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return postID, nil
}

func (repository *Repository) GetPost(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
	const postQuery = `
		SELECT
			p.post_id,
			p.author_user_id,
			p.title,
			p.required_subscription_level,
			p.created_at,
			p.updated_at,
			COALESCE(l.likes_count, 0),
			EXISTS (
				SELECT 1
				FROM content_post_like viewer_like
				WHERE viewer_like.post_id = p.post_id
					AND viewer_like.user_id = $2
			),
			COALESCE(c.comments_count, 0)
		FROM content_post p
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS likes_count
			FROM content_post_like
			GROUP BY post_id
		) l ON l.post_id = p.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comments_count
			FROM content_comment
			GROUP BY post_id
		) c ON c.post_id = p.post_id
		WHERE p.post_id = $1
	`

	var (
		post          domain.Post
		requiredLevel sql.NullInt32
	)
	err := repository.db.QueryRowContext(ctx, postQuery, postID, viewerUserID).Scan(
		&post.PostID,
		&post.AuthorUserID,
		&post.Title,
		&requiredLevel,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.LikesCount,
		&post.IsLiked,
		&post.CommentsCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Post{}, domain.ErrPostNotFound
		}
		return domain.Post{}, err
	}
	if requiredLevel.Valid {
		post.RequiredSubscriptionLevel = &requiredLevel.Int32
	}

	blocks, err := repository.listBlocks(ctx, postID)
	if err != nil {
		return domain.Post{}, err
	}
	post.Blocks = blocks

	return post, nil
}

func (repository *Repository) ListAuthorPosts(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
	const query = `
		SELECT
			p.post_id,
			p.author_user_id,
			p.title,
			p.required_subscription_level,
			p.created_at,
			COALESCE(l.likes_count, 0),
			EXISTS (
				SELECT 1
				FROM content_post_like viewer_like
				WHERE viewer_like.post_id = p.post_id
					AND viewer_like.user_id = $2
			),
			COALESCE(c.comments_count, 0)
		FROM content_post p
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS likes_count
			FROM content_post_like
			GROUP BY post_id
		) l ON l.post_id = p.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comments_count
			FROM content_comment
			GROUP BY post_id
		) c ON c.post_id = p.post_id
		WHERE p.author_user_id = $1
		ORDER BY p.created_at DESC, p.post_id DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := repository.db.QueryContext(ctx, query, authorUserID, viewerUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]domain.PostSummary, 0)
	for rows.Next() {
		var (
			post          domain.PostSummary
			requiredLevel sql.NullInt32
		)

		if err := rows.Scan(
			&post.PostID,
			&post.AuthorUserID,
			&post.Title,
			&requiredLevel,
			&post.CreatedAt,
			&post.LikesCount,
			&post.IsLiked,
			&post.CommentsCount,
		); err != nil {
			return nil, err
		}
		if requiredLevel.Valid {
			post.RequiredSubscriptionLevel = &requiredLevel.Int32
		}

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (repository *Repository) SearchPosts(ctx context.Context, searchQuery usecase.SearchPostsQuery) ([]domain.PostSummary, error) {
	const baseQuery = `
		SELECT
			p.post_id,
			p.author_user_id,
			p.title,
			p.required_subscription_level,
			p.created_at,
			COALESCE(l.likes_count, 0),
			EXISTS (
				SELECT 1
				FROM content_post_like viewer_like
				WHERE viewer_like.post_id = p.post_id
					AND viewer_like.user_id = $1
			),
			COALESCE(c.comments_count, 0)
		FROM content_post p
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS likes_count
			FROM content_post_like
			GROUP BY post_id
		) l ON l.post_id = p.post_id
		LEFT JOIN (
			SELECT post_id, COUNT(*) AS comments_count
			FROM content_comment
			GROUP BY post_id
		) c ON c.post_id = p.post_id
	`

	args := []any{searchQuery.ViewerUserID}
	conditions := make([]string, 0, 6)

	if trimmedQuery := strings.TrimSpace(searchQuery.Query); trimmedQuery != "" {
		args = append(args, "%"+trimmedQuery+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(
			conditions,
			"(p.title ILIKE "+placeholder+" OR EXISTS (SELECT 1 FROM content_post_block search_block WHERE search_block.post_id = p.post_id AND search_block.kind = 'text' AND search_block.text_content ILIKE "+placeholder+"))",
		)
	}
	if len(searchQuery.AuthorUserIDs) > 0 {
		args = append(args, pq.Array(searchQuery.AuthorUserIDs))
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, "p.author_user_id = ANY("+placeholder+")")
	}
	if len(searchQuery.BlockKinds) > 0 {
		args = append(args, pq.Array(blockKindStrings(searchQuery.BlockKinds)))
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, "EXISTS (SELECT 1 FROM content_post_block filter_block WHERE filter_block.post_id = p.post_id AND filter_block.kind::text = ANY("+placeholder+"))")
	}
	if searchQuery.MinRequiredSubscriptionLevel != nil {
		args = append(args, *searchQuery.MinRequiredSubscriptionLevel)
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, "COALESCE(p.required_subscription_level, 0) >= "+placeholder)
	}
	if searchQuery.MaxRequiredSubscriptionLevel != nil {
		args = append(args, *searchQuery.MaxRequiredSubscriptionLevel)
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, "COALESCE(p.required_subscription_level, 0) <= "+placeholder)
	}
	if searchQuery.OnlyAvailable {
		if searchQuery.ViewerSubscriptionLevel == nil {
			conditions = append(conditions, "(p.required_subscription_level IS NULL OR (p.author_user_id = $1 AND $1 > 0))")
		} else {
			args = append(args, *searchQuery.ViewerSubscriptionLevel)
			placeholder := fmt.Sprintf("$%d", len(args))
			conditions = append(conditions, "(p.required_subscription_level IS NULL OR (p.author_user_id = $1 AND $1 > 0) OR p.required_subscription_level <= "+placeholder+")")
		}
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseQuery)
	if len(conditions) > 0 {
		queryBuilder.WriteString(" WHERE ")
		queryBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	args = append(args, searchQuery.Limit, searchQuery.Offset)
	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY p.created_at DESC, p.post_id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)))

	rows, err := repository.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]domain.PostSummary, 0)
	for rows.Next() {
		var (
			post          domain.PostSummary
			requiredLevel sql.NullInt32
		)

		if err := rows.Scan(
			&post.PostID,
			&post.AuthorUserID,
			&post.Title,
			&requiredLevel,
			&post.CreatedAt,
			&post.LikesCount,
			&post.IsLiked,
			&post.CommentsCount,
		); err != nil {
			return nil, err
		}
		if requiredLevel.Valid {
			post.RequiredSubscriptionLevel = &requiredLevel.Int32
		}

		posts = append(posts, post)
	}

	return posts, rows.Err()
}

func (repository *Repository) UpdatePost(ctx context.Context, post domain.Post, replaceBlocks bool) error {
	now := time.Now().UTC()
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const updatePostQuery = `
		UPDATE content_post
		SET title = $3,
			required_subscription_level = $4,
			updated_at = $5
		WHERE post_id = $1
			AND author_user_id = $2
	`

	result, err := tx.ExecContext(
		ctx,
		updatePostQuery,
		post.PostID,
		post.AuthorUserID,
		post.Title,
		nullInt32(post.RequiredSubscriptionLevel),
		now,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ensurePostOwnership(ctx, tx, post.PostID, post.AuthorUserID)
	}

	if replaceBlocks {
		if _, err := tx.ExecContext(ctx, `DELETE FROM content_post_block WHERE post_id = $1`, post.PostID); err != nil {
			return err
		}
		if err := insertBlocks(ctx, tx, post.PostID, post.Blocks, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repository *Repository) DeletePost(ctx context.Context, postID int64, authorUserID int64) error {
	result, err := repository.db.ExecContext(
		ctx,
		`DELETE FROM content_post WHERE post_id = $1 AND author_user_id = $2`,
		postID,
		authorUserID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ensurePostOwnership(ctx, repository.db, postID, authorUserID)
	}

	return nil
}

func (repository *Repository) UpsertLike(ctx context.Context, postID int64, userID int64) error {
	now := time.Now().UTC()

	_, err := repository.db.ExecContext(
		ctx,
		`
			INSERT INTO content_post_like (post_id, user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $3)
			ON CONFLICT (post_id, user_id)
			DO UPDATE SET updated_at = EXCLUDED.updated_at
		`,
		postID,
		userID,
		now,
	)

	return err
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

type sqlQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func ensurePostOwnership(ctx context.Context, queryer sqlQueryer, postID int64, authorUserID int64) error {
	var storedAuthorUserID int64
	err := queryer.QueryRowContext(ctx, `SELECT author_user_id FROM content_post WHERE post_id = $1`, postID).Scan(&storedAuthorUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrPostNotFound
		}
		return err
	}
	if storedAuthorUserID != authorUserID {
		return domain.ErrPostForbidden
	}

	return nil
}

func (repository *Repository) listBlocks(ctx context.Context, postID int64) ([]domain.PostBlock, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
			SELECT post_block_id, position, kind, text_content, file_url
			FROM content_post_block
			WHERE post_id = $1
			ORDER BY position ASC, post_block_id ASC
		`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocks := make([]domain.PostBlock, 0)
	for rows.Next() {
		var (
			block       domain.PostBlock
			kind        string
			textContent sql.NullString
			fileURL     sql.NullString
		)
		if err := rows.Scan(&block.PostBlockID, &block.Position, &kind, &textContent, &fileURL); err != nil {
			return nil, err
		}

		block.Kind = domain.BlockKind(kind)
		if textContent.Valid {
			block.TextContent = &textContent.String
		}
		if fileURL.Valid {
			block.FileURL = &fileURL.String
		}

		blocks = append(blocks, block)
	}

	return blocks, rows.Err()
}

func insertBlocks(ctx context.Context, tx *sql.Tx, postID int64, blocks []domain.PostBlock, now time.Time) error {
	const insertBlockQuery = `
		INSERT INTO content_post_block (
			post_id,
			position,
			kind,
			text_content,
			file_url,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $6)
	`

	for _, block := range blocks {
		if _, err := tx.ExecContext(
			ctx,
			insertBlockQuery,
			postID,
			block.Position,
			string(block.Kind),
			nullString(block.TextContent),
			nullString(block.FileURL),
			now,
		); err != nil {
			return err
		}
	}

	return nil
}

func nullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *value,
		Valid:  true,
	}
}

func nullInt32(value *int32) sql.NullInt32 {
	if value == nil {
		return sql.NullInt32{}
	}

	return sql.NullInt32{
		Int32: *value,
		Valid: true,
	}
}

func blockKindStrings(kinds []domain.BlockKind) []string {
	result := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		result = append(result, string(kind))
	}

	return result
}
