package domain

type NotificationPreferences struct {
	Comments      bool
	Likes         bool
	Donations     bool
	Posts         bool
	Subscriptions bool
	Meetings      bool
	EmailDigest   bool
}

func DefaultNotificationPreferences() NotificationPreferences {
	return NotificationPreferences{
		Comments:      true,
		Likes:         true,
		Donations:     true,
		Posts:         true,
		Subscriptions: true,
		Meetings:      true,
		EmailDigest:   false,
	}
}

func (preferences NotificationPreferences) Allows(notificationType NotificationType) bool {
	switch notificationType {
	case NotificationTypeComment:
		return preferences.Comments
	case NotificationTypeLike:
		return preferences.Likes
	case NotificationTypeDonation:
		return preferences.Donations
	case NotificationTypePost:
		return preferences.Posts
	case NotificationTypeSubscription:
		return preferences.Subscriptions
	case NotificationTypeMeeting:
		return preferences.Meetings
	default:
		return true
	}
}
