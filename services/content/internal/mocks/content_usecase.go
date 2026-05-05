package mocks

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
)

type ContentUseCase struct {
	ListAuthorPostsFunc    func(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error)
	SearchPostsFunc        func(ctx context.Context, query usecase.SearchPostsQuery) ([]domain.PostSummary, error)
	CreatePostFunc         func(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error)
	UploadPostMediaFunc    func(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error)
	GetPostFunc            func(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error)
	UpdatePostFunc         func(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error)
	DeletePostFunc         func(ctx context.Context, command usecase.DeletePostCommand) error
	ListTiersFunc          func(ctx context.Context, query usecase.ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error)
	CreateTierFunc         func(ctx context.Context, command usecase.CreateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	UpdateTierFunc         func(ctx context.Context, command usecase.UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	DeleteTierFunc         func(ctx context.Context, command usecase.DeleteSubscriptionTierCommand) error
	SubscribeFunc          func(ctx context.Context, command usecase.SubscribeToTrainerCommand) (domain.Subscription, error)
	ListSubscriptionsFunc  func(ctx context.Context, query usecase.ListMySubscriptionsQuery) ([]domain.Subscription, error)
	UpdateSubscriptionFunc func(ctx context.Context, command usecase.UpdateSubscriptionCommand) (domain.Subscription, error)
	CancelSubscriptionFunc func(ctx context.Context, command usecase.CancelSubscriptionCommand) error
	LikePostFunc           func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	UnlikePostFunc         func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	CreateCommentFunc      func(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error)
	ListCommentsFunc       func(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error)
}

func (mock ContentUseCase) ListAuthorPosts(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error) {
	return mock.ListAuthorPostsFunc(ctx, query)
}

func (mock ContentUseCase) SearchPosts(ctx context.Context, query usecase.SearchPostsQuery) ([]domain.PostSummary, error) {
	return mock.SearchPostsFunc(ctx, query)
}

func (mock ContentUseCase) CreatePost(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error) {
	return mock.CreatePostFunc(ctx, command)
}

func (mock ContentUseCase) UploadPostMedia(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error) {
	return mock.UploadPostMediaFunc(ctx, command)
}

func (mock ContentUseCase) GetPost(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error) {
	return mock.GetPostFunc(ctx, query)
}

func (mock ContentUseCase) UpdatePost(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error) {
	return mock.UpdatePostFunc(ctx, command)
}

func (mock ContentUseCase) DeletePost(ctx context.Context, command usecase.DeletePostCommand) error {
	return mock.DeletePostFunc(ctx, command)
}

func (mock ContentUseCase) ListSubscriptionTiers(ctx context.Context, query usecase.ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error) {
	if mock.ListTiersFunc == nil {
		return nil, nil
	}
	return mock.ListTiersFunc(ctx, query)
}

func (mock ContentUseCase) CreateSubscriptionTier(ctx context.Context, command usecase.CreateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	if mock.CreateTierFunc == nil {
		return domain.SubscriptionTier{}, nil
	}
	return mock.CreateTierFunc(ctx, command)
}

func (mock ContentUseCase) UpdateSubscriptionTier(ctx context.Context, command usecase.UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	if mock.UpdateTierFunc == nil {
		return domain.SubscriptionTier{}, nil
	}
	return mock.UpdateTierFunc(ctx, command)
}

func (mock ContentUseCase) DeleteSubscriptionTier(ctx context.Context, command usecase.DeleteSubscriptionTierCommand) error {
	if mock.DeleteTierFunc == nil {
		return nil
	}
	return mock.DeleteTierFunc(ctx, command)
}

func (mock ContentUseCase) SubscribeToTrainer(ctx context.Context, command usecase.SubscribeToTrainerCommand) (domain.Subscription, error) {
	if mock.SubscribeFunc == nil {
		return domain.Subscription{}, nil
	}
	return mock.SubscribeFunc(ctx, command)
}

func (mock ContentUseCase) ListMySubscriptions(ctx context.Context, query usecase.ListMySubscriptionsQuery) ([]domain.Subscription, error) {
	if mock.ListSubscriptionsFunc == nil {
		return nil, nil
	}
	return mock.ListSubscriptionsFunc(ctx, query)
}

func (mock ContentUseCase) UpdateSubscription(ctx context.Context, command usecase.UpdateSubscriptionCommand) (domain.Subscription, error) {
	if mock.UpdateSubscriptionFunc == nil {
		return domain.Subscription{}, nil
	}
	return mock.UpdateSubscriptionFunc(ctx, command)
}

func (mock ContentUseCase) CancelSubscription(ctx context.Context, command usecase.CancelSubscriptionCommand) error {
	if mock.CancelSubscriptionFunc == nil {
		return nil
	}
	return mock.CancelSubscriptionFunc(ctx, command)
}

func (mock ContentUseCase) LikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
	return mock.LikePostFunc(ctx, command)
}

func (mock ContentUseCase) UnlikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error) {
	return mock.UnlikePostFunc(ctx, command)
}

func (mock ContentUseCase) CreateComment(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error) {
	return mock.CreateCommentFunc(ctx, command)
}

func (mock ContentUseCase) ListComments(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error) {
	return mock.ListCommentsFunc(ctx, query)
}
