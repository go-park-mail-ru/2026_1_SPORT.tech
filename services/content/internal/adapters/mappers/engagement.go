package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CreateCommentRequestToCommand(request *contentv1.CreateCommentRequest) usecase.CreateCommentCommand {
	return usecase.CreateCommentCommand{
		PostID:                  request.GetPostId(),
		AuthorUserID:            request.GetAuthorUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Body:                    request.GetBody(),
	}
}

func ListCommentsRequestToQuery(request *contentv1.ListCommentsRequest) usecase.ListCommentsQuery {
	return usecase.ListCommentsQuery{
		PostID:                  request.GetPostId(),
		ViewerUserID:            request.GetViewerUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func ListPostLikesRequestToQuery(request *contentv1.ListPostLikesRequest) usecase.ListPostLikesQuery {
	return usecase.ListPostLikesQuery{
		PostID:                  request.GetPostId(),
		ViewerUserID:            request.GetViewerUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func ListNotificationsRequestToQuery(request *contentv1.ListNotificationsRequest) usecase.ListNotificationsQuery {
	return usecase.ListNotificationsQuery{
		UserID: request.GetUserId(),
		Limit:  request.GetLimit(),
		Offset: request.GetOffset(),
	}
}

func MarkNotificationReadRequestToCommand(request *contentv1.MarkNotificationReadRequest) usecase.MarkNotificationReadCommand {
	return usecase.MarkNotificationReadCommand{
		UserID:         request.GetUserId(),
		NotificationID: request.GetNotificationId(),
	}
}

func NewCommentResponse(comment domain.Comment) *contentv1.CommentResponse {
	return &contentv1.CommentResponse{
		Comment: commentToProto(comment),
	}
}

func NewListCommentsResponse(comments []domain.Comment) *contentv1.ListCommentsResponse {
	response := &contentv1.ListCommentsResponse{
		Comments: make([]*contentv1.Comment, 0, len(comments)),
	}
	for _, comment := range comments {
		response.Comments = append(response.Comments, commentToProto(comment))
	}

	return response
}

func NewNotificationResponse(notification domain.Notification) *contentv1.NotificationResponse {
	return &contentv1.NotificationResponse{
		Notification: notificationToProto(notification),
	}
}

func NewListNotificationsResponse(notifications []domain.Notification) *contentv1.ListNotificationsResponse {
	response := &contentv1.ListNotificationsResponse{
		Notifications: make([]*contentv1.Notification, 0, len(notifications)),
	}
	for _, notification := range notifications {
		response.Notifications = append(response.Notifications, notificationToProto(notification))
	}

	return response
}

func NewListPostLikesResponse(likes []domain.PostLike) *contentv1.ListPostLikesResponse {
	response := &contentv1.ListPostLikesResponse{
		Likes: make([]*contentv1.PostLike, 0, len(likes)),
	}
	for _, like := range likes {
		response.Likes = append(response.Likes, postLikeToProto(like))
	}

	return response
}

func postLikeToProto(like domain.PostLike) *contentv1.PostLike {
	return &contentv1.PostLike{
		PostId:    like.PostID,
		UserId:    like.UserID,
		CreatedAt: timestamppb.New(like.CreatedAt),
	}
}

func commentToProto(comment domain.Comment) *contentv1.Comment {
	return &contentv1.Comment{
		CommentId:    comment.CommentID,
		PostId:       comment.PostID,
		AuthorUserId: comment.AuthorUserID,
		Body:         comment.Body,
		CreatedAt:    timestamppb.New(comment.CreatedAt),
		UpdatedAt:    timestamppb.New(comment.UpdatedAt),
	}
}

func NotificationPreferencesFromProto(preferences *contentv1.NotificationPreferences) domain.NotificationPreferences {
	if preferences == nil {
		return domain.DefaultNotificationPreferences()
	}

	return domain.NotificationPreferences{
		Comments:      preferences.GetComments(),
		Likes:         preferences.GetLikes(),
		Donations:     preferences.GetDonations(),
		Posts:         preferences.GetPosts(),
		Subscriptions: preferences.GetSubscriptions(),
		Meetings:      preferences.GetMeetings(),
		EmailDigest:   preferences.GetEmailDigest(),
	}
}

func NewNotificationPreferencesResponse(preferences domain.NotificationPreferences) *contentv1.NotificationPreferencesResponse {
	return &contentv1.NotificationPreferencesResponse{
		Preferences: &contentv1.NotificationPreferences{
			Comments:      preferences.Comments,
			Likes:         preferences.Likes,
			Donations:     preferences.Donations,
			Posts:         preferences.Posts,
			Subscriptions: preferences.Subscriptions,
			Meetings:      preferences.Meetings,
			EmailDigest:   preferences.EmailDigest,
		},
	}
}

func notificationToProto(notification domain.Notification) *contentv1.Notification {
	response := &contentv1.Notification{
		NotificationId: notification.NotificationID,
		UserId:         notification.UserID,
		Type:           string(notification.Type),
		ActorUserId:    notification.ActorUserID,
		Title:          notification.Title,
		Body:           notification.Body,
		IsRead:         notification.IsRead(),
		CreatedAt:      timestamppb.New(notification.CreatedAt),
		PostId:         notification.PostID,
		CommentId:      notification.CommentID,
		DonationId:     notification.DonationID,
		SubscriptionId: notification.SubscriptionID,
	}
	if notification.ReadAt != nil {
		response.ReadAt = timestamppb.New(*notification.ReadAt)
	}

	return response
}
