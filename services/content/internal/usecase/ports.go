package usecase

import (
	"context"
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

type Repositories struct {
	Posts         PostRepository
	Money         MonetizationRepository
	Engagement    EngagementRepository
	Notifications NotificationRepository
	Chat          ChatRepository
	Meeting       MeetingRepository
}

type PostRepository interface {
	CreatePost(ctx context.Context, post domain.Post) (int64, error)
	GetPost(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error)
	ListAuthorPosts(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error)
	SearchPosts(ctx context.Context, query SearchPostsQuery) ([]domain.PostSummary, error)
	UpdatePost(ctx context.Context, post domain.Post, replaceBlocks bool) error
	DeletePost(ctx context.Context, postID int64, authorUserID int64) error
}

type MonetizationRepository interface {
	ListSubscriptionTiers(ctx context.Context, trainerUserID int64) ([]domain.SubscriptionTier, error)
	GetSubscriptionTier(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error)
	CreateSubscriptionTier(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error)
	UpdateSubscriptionTier(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error)
	DeleteSubscriptionTier(ctx context.Context, trainerUserID int64, tierID int64) error
	GetActiveSubscriptionLevel(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error)
	SubscribeToTrainer(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	ListSubscriptions(ctx context.Context, clientUserID int64) ([]domain.Subscription, error)
	ListTrainerSubscribers(ctx context.Context, trainerUserID int64, limit int32, offset int32) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	CancelSubscription(ctx context.Context, clientUserID int64, subscriptionID int64) error
	CreateDonation(ctx context.Context, donation domain.Donation) (domain.Donation, error)
	CreateDonationPayment(ctx context.Context, payment domain.DonationPayment) (domain.DonationPayment, error)
	UpdateDonationPaymentProvider(ctx context.Context, paymentID int64, providerPaymentID string, confirmationURL string) (domain.DonationPayment, error)
	GetDonationPayment(ctx context.Context, senderUserID int64, paymentID int64) (domain.DonationPayment, error)
	ConfirmDonationPayment(ctx context.Context, senderUserID int64, paymentID int64, confirmationToken string) (domain.DonationPayment, error)
	GetBalance(ctx context.Context, trainerUserID int64, currency string) (domain.Balance, error)
	GetTrainerStatistics(ctx context.Context, trainerUserID int64, currency string, monthStart time.Time) (domain.TrainerStatistics, error)
	ListReceivedDonations(ctx context.Context, recipientUserID int64, limit, offset int32) ([]domain.Donation, error)
	CountReceivedDonations(ctx context.Context, recipientUserID int64) (int32, error)
}

type EngagementRepository interface {
	UpsertLike(ctx context.Context, postID int64, userID int64) (bool, error)
	DeleteLike(ctx context.Context, postID int64, userID int64) error
	GetPostLikeState(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error)
	CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	ListComments(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error)
	ListPostLikes(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.PostLike, error)
}

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification domain.Notification) (domain.Notification, error)
	ListNotifications(ctx context.Context, userID int64, limit int32, offset int32) ([]domain.Notification, error)
	MarkNotificationRead(ctx context.Context, userID int64, notificationID int64) (domain.Notification, error)
}

type ChatRepository interface {
	HasActiveChatSubscription(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error)
	IsTrainerOf(ctx context.Context, trainerUserID int64, clientUserID int64) (bool, error)
	SaveChatMessage(ctx context.Context, msg domain.ChatMessage) (domain.ChatMessage, error)
	ListChatMessages(ctx context.Context, userID int64, otherUserID int64, limit int32, offset int32) ([]domain.ChatMessage, error)
	ListChatConversations(ctx context.Context, userID int64) ([]domain.ChatConversation, error)
	MarkChatMessageRead(ctx context.Context, userID int64, messageID int64) error
}

type MeetingRepository interface {
	HasActiveCalendarSubscription(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error)
	CreateMeetingAvailabilityRule(ctx context.Context, rule domain.MeetingAvailabilityRule) (domain.MeetingAvailabilityRule, error)
	ListMeetingAvailabilityRules(ctx context.Context, trainerUserID int64) ([]domain.MeetingAvailabilityRule, error)
	DeleteMeetingAvailabilityRule(ctx context.Context, trainerUserID int64, ruleID int64) error
	CreateMeetingSlot(ctx context.Context, slot domain.MeetingSlot) (domain.MeetingSlot, error)
	DeleteMeetingSlot(ctx context.Context, trainerUserID int64, slotID int64) error
	ListMeetingSlots(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingSlot, error)
	ListConfirmedBookings(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingBooking, error)
	CountActiveClientBookings(ctx context.Context, clientUserID int64, trainerUserID int64, now time.Time) (int32, error)
	CreateBooking(ctx context.Context, booking domain.MeetingBooking) (domain.MeetingBooking, error)
	GetBooking(ctx context.Context, bookingID int64) (domain.MeetingBooking, error)
	CancelBooking(ctx context.Context, bookingID int64, cancelledByUserID int64) (domain.MeetingBooking, error)
	ListUserBookings(ctx context.Context, userID int64) ([]domain.MeetingBooking, error)
}

type PostMediaStorage interface {
	UploadPostMedia(ctx context.Context, authorUserID int64, fileName string, contentType string, file io.Reader, size int64) (string, error)
}

type PaymentProvider interface {
	ProviderName() string
	CreatePayment(ctx context.Context, request PaymentProviderCreateRequest) (PaymentProviderPayment, error)
	GetPayment(ctx context.Context, providerPaymentID string) (PaymentProviderPayment, error)
}

type PaymentProviderCreateRequest struct {
	AmountValue    int32
	Currency       string
	Description    string
	IdempotenceKey string
	ReturnURL      string
	CancelURL      string
}

type PaymentProviderPayment struct {
	ProviderPaymentID string
	Status            string
	ConfirmationURL   string
}
