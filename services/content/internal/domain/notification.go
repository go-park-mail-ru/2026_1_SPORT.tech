package domain

import "time"

type NotificationType string

const (
	NotificationTypeComment      NotificationType = "comment"
	NotificationTypeDonation     NotificationType = "donation"
	NotificationTypeLike         NotificationType = "like"
	NotificationTypePost         NotificationType = "post"
	NotificationTypeSubscription NotificationType = "subscription"
	NotificationTypeMeeting      NotificationType = "meeting"
)

type Notification struct {
	NotificationID int64
	UserID         int64
	Type           NotificationType
	ActorUserID    int64
	Title          string
	Body           string
	ReadAt         *time.Time
	CreatedAt      time.Time
	PostID         *int64
	CommentID      *int64
	DonationID     *int64
	SubscriptionID *int64
}

func (notification Notification) IsRead() bool {
	return notification.ReadAt != nil
}

func (notificationType NotificationType) IsValid() bool {
	switch notificationType {
	case NotificationTypeComment,
		NotificationTypeDonation,
		NotificationTypeLike,
		NotificationTypePost,
		NotificationTypeSubscription,
		NotificationTypeMeeting:
		return true
	default:
		return false
	}
}
