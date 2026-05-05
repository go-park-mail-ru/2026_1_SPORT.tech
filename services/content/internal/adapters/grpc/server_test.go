package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/mocks"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerGetPost(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	contentUseCase := mocks.ContentUseCase{
		ListAuthorPostsFunc: func(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error) {
			return nil, errors.New("not implemented")
		},
		CreatePostFunc: func(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error) {
			return domain.Post{}, errors.New("not implemented")
		},
		UploadPostMediaFunc: func(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error) {
			return domain.PostMedia{}, errors.New("not implemented")
		},
		GetPostFunc: func(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error) {
			return domain.Post{
				PostID:       query.PostID,
				AuthorUserID: 7,
				Title:        "Morning run",
				CreatedAt:    now,
				UpdatedAt:    now,
				CanView:      true,
			}, nil
		},
		UpdatePostFunc: func(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error) {
			return domain.Post{}, errors.New("not implemented")
		},
		DeletePostFunc: func(ctx context.Context, command usecase.DeletePostCommand) error {
			return errors.New("not implemented")
		},
		LikePostFunc: func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
			return domain.PostLikeState{}, errors.New("not implemented")
		},
		UnlikePostFunc: func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
			return domain.PostLikeState{}, errors.New("not implemented")
		},
		CreateCommentFunc: func(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error) {
			return domain.Comment{}, errors.New("not implemented")
		},
		ListCommentsFunc: func(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error) {
			return nil, errors.New("not implemented")
		},
	}
	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Posts:         contentUseCase,
		PostMedia:     contentUseCase,
		Tiers:         contentUseCase,
		Subscriptions: contentUseCase,
		Comments:      contentUseCase,
	})

	response, err := server.GetPost(context.Background(), &contentv1.GetPostRequest{PostId: 7, ViewerUserId: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetPost().GetPostId() != 7 {
		t.Fatalf("unexpected post id: %d", response.GetPost().GetPostId())
	}
}

func TestServerGetPostMapsForbidden(t *testing.T) {
	contentUseCase := mocks.ContentUseCase{
		ListAuthorPostsFunc: func(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error) {
			return nil, nil
		},
		CreatePostFunc: func(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error) {
			return domain.Post{}, nil
		},
		UploadPostMediaFunc: func(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error) {
			return domain.PostMedia{}, nil
		},
		GetPostFunc: func(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error) {
			return domain.Post{}, domain.ErrPostForbidden
		},
		UpdatePostFunc: func(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error) {
			return domain.Post{}, nil
		},
		DeletePostFunc: func(ctx context.Context, command usecase.DeletePostCommand) error { return nil },
		LikePostFunc: func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
			return domain.PostLikeState{}, nil
		},
		UnlikePostFunc: func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
			return domain.PostLikeState{}, nil
		},
		CreateCommentFunc: func(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error) {
			return domain.Comment{}, nil
		},
		ListCommentsFunc: func(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error) {
			return nil, nil
		},
	}
	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Posts:         contentUseCase,
		PostMedia:     contentUseCase,
		Tiers:         contentUseCase,
		Subscriptions: contentUseCase,
		Comments:      contentUseCase,
	})

	_, err := server.GetPost(context.Background(), &contentv1.GetPostRequest{PostId: 7, ViewerUserId: 3})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unexpected status code: %s", status.Code(err))
	}
}
