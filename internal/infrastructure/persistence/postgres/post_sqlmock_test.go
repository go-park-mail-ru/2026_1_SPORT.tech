package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
	"github.com/lib/pq"
)

func TestPostRepositoryListProfilePosts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewPostRepository(db, nil)
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"post_id", "trainer_id", "min_tier_id", "title", "created_at", "likes_count", "is_liked", "can_view"}).
		AddRow(int64(1), int64(3), nil, "free post", now, int64(4), true, true)

	mock.ExpectQuery("SELECT\\s+p\\.post_id").
		WithArgs(int64(3), int64(7)).
		WillReturnRows(rows)

	posts, err := repository.ListProfilePosts(context.Background(), 3, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 1 || posts[0].PostID != 1 {
		t.Fatalf("unexpected posts: %+v", posts)
	}
	if posts[0].LikesCount != 4 || !posts[0].IsLiked {
		t.Fatalf("unexpected likes state: %+v", posts[0])
	}
}

func TestPostRepositoryGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewPostRepository(db, nil)
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)

	postRows := sqlmock.NewRows([]string{
		"post_id", "trainer_id", "min_tier_id", "title", "text_content", "created_at", "updated_at", "likes_count", "is_liked", "can_view",
	}).AddRow(int64(10), int64(3), nil, "title", "content", now, now, int64(2), true, true)
	mock.ExpectQuery("SELECT\\s+p\\.post_id").
		WithArgs(int64(10), int64(7)).
		WillReturnRows(postRows)

	attachmentRows := sqlmock.NewRows([]string{"post_attachment_id", "kind", "file_url"}).
		AddRow(int64(1), "image", "http://example.com/file.jpg")
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT post_attachment_id, kind, file_url
		FROM post_attachment
		WHERE post_id = $1
		ORDER BY post_attachment_id
	`)).
		WithArgs(int64(10)).
		WillReturnRows(attachmentRows)

	post, err := repository.GetByID(context.Background(), 10, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.PostID != 10 || len(post.Attachments) != 1 {
		t.Fatalf("unexpected post: %+v", post)
	}
	if post.LikesCount != 2 || !post.IsLiked {
		t.Fatalf("unexpected likes state: %+v", post)
	}
}

func TestPostRepositorySetLike(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewPostRepository(db, nil)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT 1
		FROM post
		WHERE post_id = $1
	`)).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}).AddRow(1))
	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO post_like (post_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (post_id, user_id) DO UPDATE
		SET updated_at = now()
	`)).
		WithArgs(int64(10), int64(7)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM post_like
		WHERE post_id = $1
	`)).
		WithArgs(int64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(3)))
	mock.ExpectCommit()

	status, err := repository.SetLike(context.Background(), 10, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.PostID != 10 || !status.IsLiked || status.LikesCount != 3 {
		t.Fatalf("unexpected status: %+v", status)
	}
}

func TestPostRepositoryUpdateAndDelete(t *testing.T) {
	t.Run("update success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()

		repository := NewPostRepository(db, nil)
		title := "new title"
		text := "new text"

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT trainer_id, min_tier_id, title, text_content").
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"trainer_id", "min_tier_id", "title", "text_content"}).AddRow(int64(7), nil, "old title", "old text"))
		mock.ExpectExec("UPDATE post").
			WithArgs(int64(10), int64(7), nil, title, text).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = repository.Update(context.Background(), 7, 10, usecase.UpdatePostCommand{
			Title:       &title,
			TextContent: &text,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("delete forbidden", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()

		repository := NewPostRepository(db, nil)

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT trainer_id").
			WithArgs(int64(10)).
			WillReturnRows(sqlmock.NewRows([]string{"trainer_id"}).AddRow(int64(8)))
		mock.ExpectRollback()

		err = repository.Delete(context.Background(), 7, 10)
		if err != usecase.ErrPostForbidden {
			t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrPostForbidden)
		}
	})
}

func TestMapPostError(t *testing.T) {
	err := mapPostError(&pq.Error{Code: sqlStateForeignKeyViolation, Constraint: postMinTierConstraint})
	if err != usecase.ErrPostTierNotFound {
		t.Fatalf("unexpected mapped error: %v", err)
	}
}
