package usecase

import (
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

type CreatePostCommand struct {
	AuthorUserID              int64
	Title                     string
	RequiredSubscriptionLevel *int32
	SportTypeID               *int64
	Blocks                    []PostBlockInput
}

type UploadPostMediaCommand struct {
	AuthorUserID int64
	FileName     string
	ContentType  string
	Content      []byte
}

type UpdatePostCommand struct {
	PostID                         int64
	AuthorUserID                   int64
	Title                          *string
	RequiredSubscriptionLevel      *int32
	ClearRequiredSubscriptionLevel bool
	SportTypeID                    *int64
	ClearSportTypeID               bool
	Blocks                         []PostBlockInput
	ReplaceBlocks                  bool
	IsPinned                       *bool
}

type DeletePostCommand struct {
	PostID       int64
	AuthorUserID int64
}

type LikePostCommand struct {
	PostID                  int64
	UserID                  int64
	ViewerSubscriptionLevel *int32
}

type CreateCommentCommand struct {
	PostID                  int64
	AuthorUserID            int64
	ViewerSubscriptionLevel *int32
	Body                    string
}

type UpdateCommentCommand struct {
	CommentID    int64
	AuthorUserID int64
	Body         string
}

type DeleteCommentCommand struct {
	CommentID    int64
	AuthorUserID int64
}

type GetNotificationPreferencesQuery struct {
	UserID int64
}

type UpdateNotificationPreferencesCommand struct {
	UserID      int64
	Preferences domain.NotificationPreferences
}

type CreateSubscriptionTierCommand struct {
	TrainerUserID   int64
	Name            string
	Price           int32
	Description     *string
	ChatEnabled     bool
	CalendarEnabled bool
}

type UpdateSubscriptionTierCommand struct {
	TrainerUserID    int64
	TierID           int64
	Name             *string
	Price            *int32
	Description      *string
	ClearDescription bool
	ChatEnabled      *bool
	CalendarEnabled  *bool
}

type DeleteSubscriptionTierCommand struct {
	TrainerUserID int64
	TierID        int64
}

type SubscribeToTrainerCommand struct {
	ClientUserID  int64
	TrainerUserID int64
	TierID        int64
}

type UpdateSubscriptionCommand struct {
	ClientUserID   int64
	SubscriptionID int64
	TierID         int64
}

type CancelSubscriptionCommand struct {
	ClientUserID   int64
	SubscriptionID int64
}

type DonateToProfileCommand struct {
	SenderUserID    int64
	RecipientUserID int64
	AmountValue     int32
	Currency        string
	Message         *string
}

type CreateDonationPaymentCommand struct {
	SenderUserID    int64
	RecipientUserID int64
	AmountValue     int32
	Currency        string
	Message         *string
	ReturnURL       *string
	CancelURL       *string
}

type CreateSubscriptionPaymentCommand struct {
	ClientUserID  int64
	TrainerUserID int64
	TierID        int64
	ReturnURL     *string
	CancelURL     *string
}

type ConfirmDonationPaymentCommand struct {
	SenderUserID      int64
	PaymentID         int64
	ConfirmationToken string
}

type MarkNotificationReadCommand struct {
	UserID         int64
	NotificationID int64
}

type SendChatMessageCommand struct {
	SenderUserID   int64
	ReceiverUserID int64
	Body           string
}

type MarkChatMessageReadCommand struct {
	UserID    int64
	MessageID int64
}

type CreateMeetingAvailabilityRuleCommand struct {
	TrainerUserID int64
	Weekday       int32
	StartHour     int32
}

type DeleteMeetingAvailabilityRuleCommand struct {
	TrainerUserID int64
	RuleID        int64
}

type CreateMeetingSlotCommand struct {
	TrainerUserID int64
	StartsAt      time.Time
}

type DeleteMeetingSlotCommand struct {
	TrainerUserID int64
	SlotID        int64
}

type BookMeetingCommand struct {
	ClientUserID  int64
	TrainerUserID int64
	StartsAt      time.Time
}

type AssignMeetingCommand struct {
	TrainerUserID int64
	ClientUserID  int64
	StartsAt      time.Time
	DurationHours int32
	Note          *string
}

type CancelMeetingCommand struct {
	UserID    int64
	BookingID int64
}
