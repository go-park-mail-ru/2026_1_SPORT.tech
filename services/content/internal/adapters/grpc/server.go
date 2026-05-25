package grpc

import (
	"context"
	"log/slog"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
)

type PostUseCase interface {
	ListAuthorPosts(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error)
	SearchPosts(ctx context.Context, query usecase.SearchPostsQuery) ([]domain.PostSummary, error)
	CreatePost(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error)
	GetPost(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error)
	UpdatePost(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error)
	DeletePost(ctx context.Context, command usecase.DeletePostCommand) error
	LikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	UnlikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	ListPostLikes(ctx context.Context, query usecase.ListPostLikesQuery) ([]domain.PostLike, error)
}

type PostMediaUseCase interface {
	UploadPostMedia(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error)
}

type TierUseCase interface {
	ListSubscriptionTiers(ctx context.Context, query usecase.ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error)
	CreateSubscriptionTier(ctx context.Context, command usecase.CreateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	UpdateSubscriptionTier(ctx context.Context, command usecase.UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	DeleteSubscriptionTier(ctx context.Context, command usecase.DeleteSubscriptionTierCommand) error
}

type SubscriptionUseCase interface {
	SubscribeToTrainer(ctx context.Context, command usecase.SubscribeToTrainerCommand) (domain.Subscription, error)
	ListMySubscriptions(ctx context.Context, query usecase.ListMySubscriptionsQuery) ([]domain.Subscription, error)
	ListTrainerSubscribers(ctx context.Context, query usecase.ListTrainerSubscribersQuery) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, command usecase.UpdateSubscriptionCommand) (domain.Subscription, error)
	CancelSubscription(ctx context.Context, command usecase.CancelSubscriptionCommand) error
}

type CommentUseCase interface {
	CreateComment(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error)
	ListComments(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error)
}

type DonationUseCase interface {
	DonateToProfile(ctx context.Context, command usecase.DonateToProfileCommand) (domain.Donation, error)
	CreateDonationPayment(ctx context.Context, command usecase.CreateDonationPaymentCommand) (domain.DonationPayment, error)
	CreateSubscriptionPayment(ctx context.Context, command usecase.CreateSubscriptionPaymentCommand) (domain.DonationPayment, error)
	ConfirmDonationPayment(ctx context.Context, command usecase.ConfirmDonationPaymentCommand) (domain.DonationPayment, error)
	GetBalance(ctx context.Context, query usecase.GetBalanceQuery) (domain.Balance, error)
	GetTrainerStatistics(ctx context.Context, query usecase.GetTrainerStatisticsQuery) (domain.TrainerStatistics, error)
	ListReceivedDonations(ctx context.Context, query usecase.ListReceivedDonationsQuery) ([]domain.Donation, int32, error)
}

type NotificationUseCase interface {
	ListNotifications(ctx context.Context, query usecase.ListNotificationsQuery) ([]domain.Notification, error)
	MarkNotificationRead(ctx context.Context, command usecase.MarkNotificationReadCommand) (domain.Notification, error)
}

type ChatUseCase interface {
	SendChatMessage(ctx context.Context, command usecase.SendChatMessageCommand) (domain.ChatMessage, error)
	ListChatMessages(ctx context.Context, query usecase.ListChatMessagesQuery) ([]domain.ChatMessage, error)
	ListChatConversations(ctx context.Context, query usecase.ListChatConversationsQuery) ([]domain.ChatConversation, error)
	MarkChatMessageRead(ctx context.Context, command usecase.MarkChatMessageReadCommand) error
}

type MeetingUseCase interface {
	CreateMeetingAvailabilityRule(ctx context.Context, command usecase.CreateMeetingAvailabilityRuleCommand) (domain.MeetingAvailabilityRule, error)
	ListMeetingAvailabilityRules(ctx context.Context, query usecase.ListMeetingAvailabilityRulesQuery) ([]domain.MeetingAvailabilityRule, error)
	DeleteMeetingAvailabilityRule(ctx context.Context, command usecase.DeleteMeetingAvailabilityRuleCommand) error
	CreateMeetingSlot(ctx context.Context, command usecase.CreateMeetingSlotCommand) (domain.MeetingSlot, error)
	DeleteMeetingSlot(ctx context.Context, command usecase.DeleteMeetingSlotCommand) error
	ListMyMeetingSlots(ctx context.Context, query usecase.ListMyMeetingSlotsQuery) ([]domain.MeetingSlot, error)
	ListTrainerMeetingAvailability(ctx context.Context, query usecase.ListTrainerMeetingAvailabilityQuery) ([]domain.MeetingAvailabilitySlot, error)
	BookMeeting(ctx context.Context, command usecase.BookMeetingCommand) (domain.MeetingBooking, error)
	AssignMeeting(ctx context.Context, command usecase.AssignMeetingCommand) (domain.MeetingBooking, error)
	CancelMeeting(ctx context.Context, command usecase.CancelMeetingCommand) error
	ListMeetings(ctx context.Context, query usecase.ListMeetingsQuery) ([]domain.MeetingBooking, error)
}

type UseCases struct {
	Posts         PostUseCase
	PostMedia     PostMediaUseCase
	Tiers         TierUseCase
	Subscriptions SubscriptionUseCase
	Comments      CommentUseCase
	Donations     DonationUseCase
	Notifications NotificationUseCase
	Chat          ChatUseCase
	Meeting       MeetingUseCase
}

type Server struct {
	contentv1.UnimplementedContentServiceServer
	useCases UseCases
	logger   *slog.Logger
}

func NewServer(useCases UseCases, loggers ...*slog.Logger) *Server {
	logger := slog.Default()
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	}

	return &Server{useCases: useCases, logger: logger}
}

func (server *Server) statusError(method string, err error) error {
	server.logger.Error("content grpc method failed", "method", method, "error", err)
	return mappers.ErrorToStatus(err)
}
