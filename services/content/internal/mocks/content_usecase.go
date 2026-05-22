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
	ListSubscribersFunc    func(ctx context.Context, query usecase.ListTrainerSubscribersQuery) ([]domain.Subscription, error)
	UpdateSubscriptionFunc func(ctx context.Context, command usecase.UpdateSubscriptionCommand) (domain.Subscription, error)
	CancelSubscriptionFunc func(ctx context.Context, command usecase.CancelSubscriptionCommand) error
	DonateFunc             func(ctx context.Context, command usecase.DonateToProfileCommand) (domain.Donation, error)
	CreatePaymentFunc      func(ctx context.Context, command usecase.CreateDonationPaymentCommand) (domain.DonationPayment, error)
	CreateSubPaymentFunc   func(ctx context.Context, command usecase.CreateSubscriptionPaymentCommand) (domain.DonationPayment, error)
	ConfirmPaymentFunc     func(ctx context.Context, command usecase.ConfirmDonationPaymentCommand) (domain.DonationPayment, error)
	GetBalanceFunc         func(ctx context.Context, query usecase.GetBalanceQuery) (domain.Balance, error)
	GetStatisticsFunc      func(ctx context.Context, query usecase.GetTrainerStatisticsQuery) (domain.TrainerStatistics, error)
	LikePostFunc           func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	UnlikePostFunc         func(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	CreateCommentFunc      func(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error)
	ListCommentsFunc       func(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error)
	ListNotificationsFunc  func(ctx context.Context, query usecase.ListNotificationsQuery) ([]domain.Notification, error)
	MarkNotificationFunc   func(ctx context.Context, command usecase.MarkNotificationReadCommand) (domain.Notification, error)
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

func (mock ContentUseCase) ListTrainerSubscribers(ctx context.Context, query usecase.ListTrainerSubscribersQuery) ([]domain.Subscription, error) {
	if mock.ListSubscribersFunc == nil {
		return nil, nil
	}
	return mock.ListSubscribersFunc(ctx, query)
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

func (mock ContentUseCase) DonateToProfile(ctx context.Context, command usecase.DonateToProfileCommand) (domain.Donation, error) {
	if mock.DonateFunc == nil {
		return domain.Donation{}, nil
	}
	return mock.DonateFunc(ctx, command)
}

func (mock ContentUseCase) CreateDonationPayment(ctx context.Context, command usecase.CreateDonationPaymentCommand) (domain.DonationPayment, error) {
	if mock.CreatePaymentFunc == nil {
		return domain.DonationPayment{}, nil
	}
	return mock.CreatePaymentFunc(ctx, command)
}

func (mock ContentUseCase) CreateSubscriptionPayment(ctx context.Context, command usecase.CreateSubscriptionPaymentCommand) (domain.DonationPayment, error) {
	if mock.CreateSubPaymentFunc == nil {
		return domain.DonationPayment{}, nil
	}
	return mock.CreateSubPaymentFunc(ctx, command)
}

func (mock ContentUseCase) ConfirmDonationPayment(ctx context.Context, command usecase.ConfirmDonationPaymentCommand) (domain.DonationPayment, error) {
	if mock.ConfirmPaymentFunc == nil {
		return domain.DonationPayment{}, nil
	}
	return mock.ConfirmPaymentFunc(ctx, command)
}

func (mock ContentUseCase) GetBalance(ctx context.Context, query usecase.GetBalanceQuery) (domain.Balance, error) {
	if mock.GetBalanceFunc == nil {
		return domain.Balance{}, nil
	}
	return mock.GetBalanceFunc(ctx, query)
}

func (mock ContentUseCase) GetTrainerStatistics(ctx context.Context, query usecase.GetTrainerStatisticsQuery) (domain.TrainerStatistics, error) {
	if mock.GetStatisticsFunc == nil {
		return domain.TrainerStatistics{}, nil
	}
	return mock.GetStatisticsFunc(ctx, query)
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

func (mock ContentUseCase) ListNotifications(ctx context.Context, query usecase.ListNotificationsQuery) ([]domain.Notification, error) {
	if mock.ListNotificationsFunc == nil {
		return nil, nil
	}
	return mock.ListNotificationsFunc(ctx, query)
}

func (mock ContentUseCase) MarkNotificationRead(ctx context.Context, command usecase.MarkNotificationReadCommand) (domain.Notification, error) {
	if mock.MarkNotificationFunc == nil {
		return domain.Notification{}, nil
	}
	return mock.MarkNotificationFunc(ctx, command)
}
