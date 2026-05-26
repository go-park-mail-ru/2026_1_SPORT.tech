package grpc

import (
	"context"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
)

func (server *Server) ListMyNotifications(ctx context.Context, request *gatewayv1.ListNotificationsRequest) (*gatewayv1.ListNotificationsResponse, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListNotifications(
		forwardContext(ctx),
		mappers.ListNotificationsRequestToContent(userID, request),
	)
	if err != nil {
		return nil, err
	}

	return mappers.ListNotificationsResponseFromContent(response)
}

func (server *Server) MarkNotificationRead(ctx context.Context, request *gatewayv1.MarkNotificationReadRequest) (*gatewayv1.NotificationResponse, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.MarkNotificationRead(
		forwardContext(ctx),
		mappers.MarkNotificationReadRequestToContent(userID, request),
	)
	if err != nil {
		return nil, err
	}

	return mappers.NotificationResponseFromContent(response)
}
