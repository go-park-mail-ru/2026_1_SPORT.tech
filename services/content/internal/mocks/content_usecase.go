//go:generate mockgen -source=$GOFILE -destination=content_usecase_mock.go -package=mocks

package mocks

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
)

type ContentUseCase interface {
	ListAuthorPosts(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error)
	SearchPosts(ctx context.Context, query usecase.SearchPostsQuery) ([]domain.PostSummary, error)
	CreatePost(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error)
	UploadPostMedia(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error)
	GetPost(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error)
	UpdatePost(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error)
	DeletePost(ctx context.Context, command usecase.DeletePostCommand) error
	ListSubscriptionTiers(ctx context.Context, query usecase.ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error)
	CreateSubscriptionTier(ctx context.Context, command usecase.CreateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	UpdateSubscriptionTier(ctx context.Context, command usecase.UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	DeleteSubscriptionTier(ctx context.Context, command usecase.DeleteSubscriptionTierCommand) error
	SubscribeToTrainer(ctx context.Context, command usecase.SubscribeToTrainerCommand) (domain.Subscription, error)
	ListMySubscriptions(ctx context.Context, query usecase.ListMySubscriptionsQuery) ([]domain.Subscription, error)
	ListTrainerSubscribers(ctx context.Context, query usecase.ListTrainerSubscribersQuery) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, command usecase.UpdateSubscriptionCommand) (domain.Subscription, error)
	CancelSubscription(ctx context.Context, command usecase.CancelSubscriptionCommand) error
	DonateToProfile(ctx context.Context, command usecase.DonateToProfileCommand) (domain.Donation, error)
	CreateDonationPayment(ctx context.Context, command usecase.CreateDonationPaymentCommand) (domain.DonationPayment, error)
	CreateSubscriptionPayment(ctx context.Context, command usecase.CreateSubscriptionPaymentCommand) (domain.DonationPayment, error)
	ConfirmDonationPayment(ctx context.Context, command usecase.ConfirmDonationPaymentCommand) (domain.DonationPayment, error)
	GetBalance(ctx context.Context, query usecase.GetBalanceQuery) (domain.Balance, error)
	GetTrainerStatistics(ctx context.Context, query usecase.GetTrainerStatisticsQuery) (domain.TrainerStatistics, error)
	ListReceivedDonations(ctx context.Context, query usecase.ListReceivedDonationsQuery) ([]domain.Donation, int32, error)
	LikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	UnlikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	CreateComment(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error)
	ListComments(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error)
	UpdateComment(ctx context.Context, command usecase.UpdateCommentCommand) (domain.Comment, error)
	DeleteComment(ctx context.Context, command usecase.DeleteCommentCommand) error
	ListPostLikes(ctx context.Context, query usecase.ListPostLikesQuery) ([]domain.PostLike, error)
	ListNotifications(ctx context.Context, query usecase.ListNotificationsQuery) ([]domain.Notification, error)
	MarkNotificationRead(ctx context.Context, command usecase.MarkNotificationReadCommand) (domain.Notification, error)
	GetNotificationPreferences(ctx context.Context, query usecase.GetNotificationPreferencesQuery) (domain.NotificationPreferences, error)
	UpdateNotificationPreferences(ctx context.Context, command usecase.UpdateNotificationPreferencesCommand) (domain.NotificationPreferences, error)
}
