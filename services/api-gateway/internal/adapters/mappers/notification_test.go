package mappers

import (
	"testing"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func TestListNotificationsRequestToContent(t *testing.T) {
	req := ListNotificationsRequestToContent(7, &gatewayv1.ListNotificationsRequest{
		Limit:  10,
		Offset: 5,
	})
	if req.GetUserId() != 7 || req.GetLimit() != 10 || req.GetOffset() != 5 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestMarkNotificationReadRequestToContent(t *testing.T) {
	req := MarkNotificationReadRequestToContent(3, &gatewayv1.MarkNotificationReadRequest{
		NotificationId: 42,
	})
	if req.GetUserId() != 3 || req.GetNotificationId() != 42 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestNotificationResponseFromContent(t *testing.T) {
	resp, err := NotificationResponseFromContent(&contentv1.NotificationResponse{
		Notification: &contentv1.Notification{
			NotificationId: 1,
			UserId:         7,
			ActorUserId:    8,
			Type:           "new_post",
			Title:          "New post",
			Body:           "Check it out",
			IsRead:         false,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetNotification().GetNotificationId() != 1 {
		t.Fatalf("unexpected notification id: %d", resp.GetNotification().GetNotificationId())
	}
	if resp.GetNotification().GetType() != "new_post" {
		t.Fatalf("unexpected notification type: %s", resp.GetNotification().GetType())
	}
}

func TestNotificationResponseFromContentNil(t *testing.T) {
	_, err := NotificationResponseFromContent(nil)
	if err == nil {
		t.Fatal("expected error for nil response")
	}

	_, err = NotificationResponseFromContent(&contentv1.NotificationResponse{})
	if err == nil {
		t.Fatal("expected error for nil notification")
	}
}

func TestListNotificationsResponseFromContent(t *testing.T) {
	resp, err := ListNotificationsResponseFromContent(&contentv1.ListNotificationsResponse{
		Notifications: []*contentv1.Notification{
			{NotificationId: 1, UserId: 7, ActorUserId: 8, Type: "new_post"},
			{NotificationId: 2, UserId: 7, ActorUserId: 9, Type: "donation"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetNotifications()) != 2 {
		t.Fatalf("unexpected notifications count: %d", len(resp.GetNotifications()))
	}
}

func TestListNotificationsResponseFromContentNil(t *testing.T) {
	_, err := ListNotificationsResponseFromContent(nil)
	if err == nil {
		t.Fatal("expected error for nil response")
	}
}

func TestListNotificationsResponseFromContentEmpty(t *testing.T) {
	resp, err := ListNotificationsResponseFromContent(&contentv1.ListNotificationsResponse{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetNotifications()) != 0 {
		t.Fatalf("unexpected notifications: %+v", resp.GetNotifications())
	}
}
