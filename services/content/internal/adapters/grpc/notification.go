package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
)

func (server *Server) ListNotifications(ctx context.Context, request *contentv1.ListNotificationsRequest) (*contentv1.ListNotificationsResponse, error) {
	notifications, err := server.useCases.Notifications.ListNotifications(ctx, mappers.ListNotificationsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListNotificationsResponse(notifications), nil
}

func (server *Server) MarkNotificationRead(ctx context.Context, request *contentv1.MarkNotificationReadRequest) (*contentv1.NotificationResponse, error) {
	notification, err := server.useCases.Notifications.MarkNotificationRead(ctx, mappers.MarkNotificationReadRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewNotificationResponse(notification), nil
}
