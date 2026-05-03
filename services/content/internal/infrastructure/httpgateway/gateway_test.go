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
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
)

type stubContentUseCase struct {
	listAuthorPostsFunc    func(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error)
	searchPostsFunc        func(ctx context.Context, query usecase.SearchPostsQuery) ([]domain.PostSummary, error)
	createPostFunc         func(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error)
	uploadPostMediaFunc    func(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error)
	getPostFunc            func(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error)
	updatePostFunc         func(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error)
	deletePostFunc         func(ctx context.Context, command usecase.DeletePostCommand) error
	listTiersFunc          func(ctx context.Context, query usecase.ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error)
	createTierFunc         func(ctx context.Context, command usecase.CreateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	updateTierFunc         func(ctx context.Context, command usecase.UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	deleteTierFunc         func(ctx context.Context, command usecase.DeleteSubscriptionTierCommand) error
	subscribeFunc          func(ctx context.Context, command usecase.SubscribeToTrainerCommand) (domain.Subscription, error)
	listSubscriptionsFunc  func(ctx context.Context, query usecase.ListMySubscriptionsQuery) ([]domain.Subscription, error)
	cancelSubscriptionFunc func(ctx context.Context, command usecase.CancelSubscriptionCommand) error
	likePostFunc           func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	unlikePostFunc         func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	createCommentFunc      func(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error)
	listCommentsFunc       func(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error)
}

func (stub stubContentUseCase) ListAuthorPosts(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error) {
	return stub.listAuthorPostsFunc(ctx, query)
}

func (stub stubContentUseCase) SearchPosts(ctx context.Context, query usecase.SearchPostsQuery) ([]domain.PostSummary, error) {
	return stub.searchPostsFunc(ctx, query)
}

func (stub stubContentUseCase) CreatePost(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error) {
	return stub.createPostFunc(ctx, command)
}

func (stub stubContentUseCase) UploadPostMedia(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error) {
	return stub.uploadPostMediaFunc(ctx, command)
}

func (stub stubContentUseCase) GetPost(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error) {
	return stub.getPostFunc(ctx, query)
}

func (stub stubContentUseCase) UpdatePost(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error) {
	return stub.updatePostFunc(ctx, command)
}

func (stub stubContentUseCase) DeletePost(ctx context.Context, command usecase.DeletePostCommand) error {
	return stub.deletePostFunc(ctx, command)
}

func (stub stubContentUseCase) ListSubscriptionTiers(ctx context.Context, query usecase.ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error) {
	if stub.listTiersFunc == nil {
		return nil, nil
	}
	return stub.listTiersFunc(ctx, query)
}

func (stub stubContentUseCase) CreateSubscriptionTier(ctx context.Context, command usecase.CreateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	if stub.createTierFunc == nil {
		return domain.SubscriptionTier{}, nil
	}
	return stub.createTierFunc(ctx, command)
}

func (stub stubContentUseCase) UpdateSubscriptionTier(ctx context.Context, command usecase.UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	if stub.updateTierFunc == nil {
		return domain.SubscriptionTier{}, nil
	}
	return stub.updateTierFunc(ctx, command)
}

func (stub stubContentUseCase) DeleteSubscriptionTier(ctx context.Context, command usecase.DeleteSubscriptionTierCommand) error {
	if stub.deleteTierFunc == nil {
		return nil
	}
	return stub.deleteTierFunc(ctx, command)
}

func (stub stubContentUseCase) SubscribeToTrainer(ctx context.Context, command usecase.SubscribeToTrainerCommand) (domain.Subscription, error) {
	if stub.subscribeFunc == nil {
		return domain.Subscription{}, nil
	}
	return stub.subscribeFunc(ctx, command)
}

func (stub stubContentUseCase) ListMySubscriptions(ctx context.Context, query usecase.ListMySubscriptionsQuery) ([]domain.Subscription, error) {
	if stub.listSubscriptionsFunc == nil {
		return nil, nil
	}
	return stub.listSubscriptionsFunc(ctx, query)
}

func (stub stubContentUseCase) CancelSubscription(ctx context.Context, command usecase.CancelSubscriptionCommand) error {
	if stub.cancelSubscriptionFunc == nil {
		return nil
	}
	return stub.cancelSubscriptionFunc(ctx, command)
}

func (stub stubContentUseCase) LikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
	return stub.likePostFunc(ctx, command)
}

func (stub stubContentUseCase) UnlikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
	return stub.unlikePostFunc(ctx, command)
}

func (stub stubContentUseCase) CreateComment(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error) {
	return stub.createCommentFunc(ctx, command)
}

func (stub stubContentUseCase) ListComments(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error) {
	return stub.listCommentsFunc(ctx, query)
}

func TestNewLocalMuxExposesGeneratedGetPostEndpoint(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	handler := grpcadapter.NewServer(stubContentUseCase{
		listAuthorPostsFunc: func(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error) {
			return nil, errors.New("not implemented")
		},
		createPostFunc: func(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error) {
			return domain.Post{}, errors.New("not implemented")
		},
		uploadPostMediaFunc: func(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error) {
			return domain.PostMedia{}, errors.New("not implemented")
		},
		getPostFunc: func(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error) {
			return domain.Post{
				PostID:       query.PostID,
				AuthorUserID: 7,
				Title:        "Morning run",
				CreatedAt:    now,
				UpdatedAt:    now,
				CanView:      true,
			}, nil
		},
		updatePostFunc: func(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error) {
			return domain.Post{}, errors.New("not implemented")
		},
		deletePostFunc: func(ctx context.Context, command usecase.DeletePostCommand) error {
			return errors.New("not implemented")
		},
		likePostFunc: func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
			return domain.PostLikeState{}, errors.New("not implemented")
		},
		unlikePostFunc: func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
			return domain.PostLikeState{}, errors.New("not implemented")
		},
		createCommentFunc: func(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error) {
			return domain.Comment{}, errors.New("not implemented")
		},
		listCommentsFunc: func(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error) {
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
