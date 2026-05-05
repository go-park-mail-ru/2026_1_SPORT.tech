package httpgateway_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/httpgateway"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/mocks"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
)

func TestNewLocalMuxExposesGeneratedGetPostEndpoint(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	handler := grpcadapter.NewServer(mocks.ContentUseCase{
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
	})

	mux, err := httpgateway.NewLocalMux(context.Background(), handler)
	if err != nil {
		t.Fatalf("new local mux: %v", err)
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	response, err := http.Get(server.URL + "/v1/posts/7?viewerUserId=7")
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", response.StatusCode)
	}

	var payload struct {
		Post struct {
			PostID string `json:"postId"`
		} `json:"post"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Post.PostID != "7" {
		t.Fatalf("unexpected post id: %s", payload.Post.PostID)
	}
}
