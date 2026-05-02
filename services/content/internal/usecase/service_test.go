package usecase

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

type stubContentRepository struct {
	createPostFunc      func(ctx context.Context, post domain.Post) (int64, error)
	getPostFunc         func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error)
	listAuthorPostsFunc func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error)
	updatePostFunc      func(ctx context.Context, post domain.Post, replaceBlocks bool) error
	deletePostFunc      func(ctx context.Context, postID int64, authorUserID int64) error
	upsertLikeFunc      func(ctx context.Context, postID int64, userID int64) error
	deleteLikeFunc      func(ctx context.Context, postID int64, userID int64) error
	getLikeStateFunc    func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error)
	createCommentFunc   func(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	listCommentsFunc    func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error)
}

func (repository stubContentRepository) CreatePost(ctx context.Context, post domain.Post) (int64, error) {
	return repository.createPostFunc(ctx, post)
}

func (repository stubContentRepository) GetPost(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
	return repository.getPostFunc(ctx, postID, viewerUserID)
}

func (repository stubContentRepository) ListAuthorPosts(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
	return repository.listAuthorPostsFunc(ctx, authorUserID, viewerUserID, limit, offset)
}

func (repository stubContentRepository) UpdatePost(ctx context.Context, post domain.Post, replaceBlocks bool) error {
	return repository.updatePostFunc(ctx, post, replaceBlocks)
}

func (repository stubContentRepository) DeletePost(ctx context.Context, postID int64, authorUserID int64) error {
	return repository.deletePostFunc(ctx, postID, authorUserID)
}

func (repository stubContentRepository) UpsertLike(ctx context.Context, postID int64, userID int64) error {
	return repository.upsertLikeFunc(ctx, postID, userID)
}

func (repository stubContentRepository) DeleteLike(ctx context.Context, postID int64, userID int64) error {
	return repository.deleteLikeFunc(ctx, postID, userID)
}

func (repository stubContentRepository) GetPostLikeState(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
	return repository.getLikeStateFunc(ctx, postID, userID)
}

func (repository stubContentRepository) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	return repository.createCommentFunc(ctx, comment)
}

func (repository stubContentRepository) ListComments(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
	return repository.listCommentsFunc(ctx, postID, limit, offset)
}

type stubPostMediaStorage struct {
	uploadFunc func(ctx context.Context, authorUserID int64, fileName string, contentType string, file io.Reader, size int64) (string, error)
}

func (storage stubPostMediaStorage) UploadPostMedia(
	ctx context.Context,
	authorUserID int64,
	fileName string,
	contentType string,
	file io.Reader,
	size int64,
) (string, error) {
	return storage.uploadFunc(ctx, authorUserID, fileName, contentType, file, size)
}

func TestServiceCreatePost(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	requiredLevel := int32(2)

	service := NewService(
		stubContentRepository{
			createPostFunc: func(ctx context.Context, post domain.Post) (int64, error) {
				if post.AuthorUserID != 7 || post.Title != "Morning run" || len(post.Blocks) != 2 {
					t.Fatalf("unexpected post: %+v", post)
				}
				return 101, nil
			},
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{
					PostID:                    postID,
					AuthorUserID:              viewerUserID,
					Title:                     "Morning run",
					RequiredSubscriptionLevel: &requiredLevel,
					CreatedAt:                 now,
					UpdatedAt:                 now,
					Blocks: []domain.PostBlock{{
						PostBlockID: 1,
						Position:    0,
						Kind:        domain.BlockKindText,
						TextContent: stringPtr("Warm-up"),
					}},
				}, nil
			},
			listAuthorPostsFunc: func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
				return nil, nil
			},
			updatePostFunc: func(ctx context.Context, post domain.Post, replaceBlocks bool) error { return nil },
			deletePostFunc: func(ctx context.Context, postID int64, authorUserID int64) error { return nil },
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			deleteLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			getLikeStateFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
				return domain.PostLikeState{}, nil
			},
			createCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
				return domain.Comment{}, nil
			},
			listCommentsFunc: func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
				return nil, nil
			},
		},
		nil,
	)

	post, err := service.CreatePost(context.Background(), CreatePostCommand{
		AuthorUserID:              7,
		Title:                     " Morning run ",
		RequiredSubscriptionLevel: &requiredLevel,
		Blocks: []PostBlockInput{
			{Kind: domain.BlockKindText, TextContent: stringPtr(" Warm-up ")},
			{Kind: domain.BlockKindImage, FileURL: stringPtr(" https://cdn.example/run.jpg ")},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.PostID != 101 {
		t.Fatalf("unexpected post id: %d", post.PostID)
	}
}

func TestServiceUploadPostMedia(t *testing.T) {
	uploaded := false
	service := NewService(
		stubContentRepository{},
		stubPostMediaStorage{
			uploadFunc: func(ctx context.Context, authorUserID int64, fileName string, contentType string, file io.Reader, size int64) (string, error) {
				uploaded = true
				if authorUserID != 7 || fileName != "run.png" || contentType != "image/png" || size != 4 {
					t.Fatalf("unexpected upload args: authorUserID=%d fileName=%s contentType=%s size=%d", authorUserID, fileName, contentType, size)
				}

				content, err := io.ReadAll(file)
				if err != nil {
					t.Fatalf("read upload content: %v", err)
				}
				if string(content) != "data" {
					t.Fatalf("unexpected upload content: %q", string(content))
				}

				return "http://storage/posts/7/run.png", nil
			},
		},
	)

	media, err := service.UploadPostMedia(context.Background(), UploadPostMediaCommand{
		AuthorUserID: 7,
		FileName:     " run.png ",
		ContentType:  " image/png ",
		Content:      []byte("data"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !uploaded {
		t.Fatal("expected upload to be called")
	}
	if media.FileURL != "http://storage/posts/7/run.png" ||
		media.Kind != domain.BlockKindImage ||
		media.ContentType != "image/png" ||
		media.SizeBytes != 4 {
		t.Fatalf("unexpected media: %+v", media)
	}
}

func TestServiceGetPostRejectsRestrictedAccess(t *testing.T) {
	requiredLevel := int32(2)

	service := NewService(
		stubContentRepository{
			createPostFunc: func(ctx context.Context, post domain.Post) (int64, error) { return 0, nil },
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{
					PostID:                    postID,
					AuthorUserID:              9,
					Title:                     "Subscribers only",
					RequiredSubscriptionLevel: &requiredLevel,
				}, nil
			},
			listAuthorPostsFunc: func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
				return nil, nil
			},
			updatePostFunc: func(ctx context.Context, post domain.Post, replaceBlocks bool) error { return nil },
			deletePostFunc: func(ctx context.Context, postID int64, authorUserID int64) error { return nil },
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			deleteLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			getLikeStateFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
				return domain.PostLikeState{}, nil
			},
			createCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
				return domain.Comment{}, nil
			},
			listCommentsFunc: func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
				return nil, nil
			},
		},
		nil,
	)

	_, err := service.GetPost(context.Background(), GetPostQuery{
		PostID:       33,
		ViewerUserID: 7,
	})
	if !errors.Is(err, domain.ErrPostForbidden) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreateComment(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)

	service := NewService(
		stubContentRepository{
			createPostFunc: func(ctx context.Context, post domain.Post) (int64, error) { return 0, nil },
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{
					PostID:       postID,
					AuthorUserID: 7,
					Title:        "Public post",
				}, nil
			},
			listAuthorPostsFunc: func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
				return nil, nil
			},
			updatePostFunc: func(ctx context.Context, post domain.Post, replaceBlocks bool) error { return nil },
			deletePostFunc: func(ctx context.Context, postID int64, authorUserID int64) error { return nil },
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			deleteLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			getLikeStateFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
				return domain.PostLikeState{}, nil
			},
			createCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
				if comment.PostID != 21 || comment.AuthorUserID != 13 || comment.Body != "Great workout" {
					t.Fatalf("unexpected comment: %+v", comment)
				}
				comment.CommentID = 88
				comment.CreatedAt = now
				comment.UpdatedAt = now
				return comment, nil
			},
			listCommentsFunc: func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
				return nil, nil
			},
		},
		nil,
	)

	comment, err := service.CreateComment(context.Background(), CreateCommentCommand{
		PostID:       21,
		AuthorUserID: 13,
		Body:         " Great workout ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.CommentID != 88 {
		t.Fatalf("unexpected comment id: %d", comment.CommentID)
	}
}

func stringPtr(value string) *string {
	return &value
}
