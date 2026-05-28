package mappers

import (
	"fmt"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func NotificationPreferencesToContent(preferences *gatewayv1.NotificationPreferences) *contentv1.NotificationPreferences {
	if preferences == nil {
		return nil
	}

	return &contentv1.NotificationPreferences{
		Comments:      preferences.GetComments(),
		Likes:         preferences.GetLikes(),
		Donations:     preferences.GetDonations(),
		Posts:         preferences.GetPosts(),
		Subscriptions: preferences.GetSubscriptions(),
		Meetings:      preferences.GetMeetings(),
		EmailDigest:   preferences.GetEmailDigest(),
	}
}

func NotificationPreferencesResponseFromContent(response *contentv1.NotificationPreferencesResponse) *gatewayv1.NotificationPreferencesResponse {
	preferences := response.GetPreferences()

	return &gatewayv1.NotificationPreferencesResponse{
		Preferences: &gatewayv1.NotificationPreferences{
			Comments:      preferences.GetComments(),
			Likes:         preferences.GetLikes(),
			Donations:     preferences.GetDonations(),
			Posts:         preferences.GetPosts(),
			Subscriptions: preferences.GetSubscriptions(),
			Meetings:      preferences.GetMeetings(),
			EmailDigest:   preferences.GetEmailDigest(),
		},
	}
}

func ListNotificationsRequestToContent(userID int64, request *gatewayv1.ListNotificationsRequest) *contentv1.ListNotificationsRequest {
	return &contentv1.ListNotificationsRequest{
		UserId: userID,
		Limit:  request.GetLimit(),
		Offset: request.GetOffset(),
	}
}

func MarkNotificationReadRequestToContent(userID int64, request *gatewayv1.MarkNotificationReadRequest) *contentv1.MarkNotificationReadRequest {
	return &contentv1.MarkNotificationReadRequest{
		UserId:         userID,
		NotificationId: int32ToInt64(request.GetNotificationId()),
	}
}

func NotificationResponseFromContent(response *contentv1.NotificationResponse) (*gatewayv1.NotificationResponse, error) {
	if response == nil || response.GetNotification() == nil {
		return nil, fmt.Errorf("notification is required")
	}

	notification, err := notificationFromContent(response.GetNotification())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.NotificationResponse{Notification: notification}, nil
}

func ListNotificationsResponseFromContent(response *contentv1.ListNotificationsResponse) (*gatewayv1.ListNotificationsResponse, error) {
	if response == nil {
		return nil, fmt.Errorf("notifications response is required")
	}

	result := &gatewayv1.ListNotificationsResponse{
		Notifications: make([]*gatewayv1.Notification, 0, len(response.GetNotifications())),
	}
	for _, item := range response.GetNotifications() {
		notification, err := notificationFromContent(item)
		if err != nil {
			return nil, err
		}
		result.Notifications = append(result.Notifications, notification)
	}

	return result, nil
}

func notificationFromContent(notification *contentv1.Notification) (*gatewayv1.Notification, error) {
	if notification == nil {
		return nil, fmt.Errorf("notification is required")
	}

	notificationID, err := int64ToInt32("content.notification.notification_id", notification.GetNotificationId())
	if err != nil {
		return nil, err
	}
	userID, err := int64ToInt32("content.notification.user_id", notification.GetUserId())
	if err != nil {
		return nil, err
	}
	actorUserID, err := int64ToInt32("content.notification.actor_user_id", notification.GetActorUserId())
	if err != nil {
		return nil, err
	}
	postID, err := optionalInt64ToInt32("content.notification.post_id", notification.PostId)
	if err != nil {
		return nil, err
	}
	commentID, err := optionalInt64ToInt32("content.notification.comment_id", notification.CommentId)
	if err != nil {
		return nil, err
	}
	donationID, err := optionalInt64ToInt32("content.notification.donation_id", notification.DonationId)
	if err != nil {
		return nil, err
	}
	subscriptionID, err := optionalInt64ToInt32("content.notification.subscription_id", notification.SubscriptionId)
	if err != nil {
		return nil, err
	}

	return &gatewayv1.Notification{
		NotificationId: notificationID,
		UserId:         userID,
		Type:           notification.GetType(),
		ActorUserId:    actorUserID,
		Title:          notification.GetTitle(),
		Body:           notification.GetBody(),
		IsRead:         notification.GetIsRead(),
		CreatedAt:      notification.GetCreatedAt(),
		ReadAt:         notification.ReadAt,
		PostId:         postID,
		CommentId:      commentID,
		DonationId:     donationID,
		SubscriptionId: subscriptionID,
	}, nil
}
