package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) SubscribeToTrainer(ctx context.Context, request *gatewayv1.SubscribeRequest) (*gatewayv1.Subscription, error) {
	return nil, status.Error(codes.FailedPrecondition, "subscription payment is required")
}

func (server *Server) ListMySubscriptions(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.SubscriptionsResponse, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListMySubscriptions(
		forwardContext(ctx),
		&contentv1.ListMySubscriptionsRequest{ClientUserId: userID},
	)
	if err != nil {
		return nil, err
	}

	return mappers.SubscriptionsResponseFromContent(response)
}

func (server *Server) ListMySubscribers(ctx context.Context, request *gatewayv1.ListSubscribersRequest) (*gatewayv1.SubscribersResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	if err := mappers.RequireTrainerRole(principal.User); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	trainerUserID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.contentClient.ListTrainerSubscribers(
		forwardContext(ctx),
		&contentv1.ListTrainerSubscribersRequest{
			TrainerUserId: trainerUserID,
			Limit:         request.GetLimit(),
			Offset:        request.GetOffset(),
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.SubscribersResponseFromContent(response)
}

func (server *Server) UpdateSubscription(ctx context.Context, request *gatewayv1.UpdateSubscriptionRequest) (*gatewayv1.Subscription, error) {
	return nil, status.Error(codes.FailedPrecondition, "subscription payment is required")
}

func (server *Server) CancelSubscription(ctx context.Context, request *gatewayv1.CancelSubscriptionRequest) (*emptypb.Empty, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := server.contentClient.CancelSubscription(
		forwardContext(ctx),
		&contentv1.CancelSubscriptionRequest{
			ClientUserId:   userID,
			SubscriptionId: int64(request.GetSubscriptionId()),
		},
	); err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 204); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) requireSubscriptionUserID(ctx context.Context) (int64, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return 0, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "unauthorized")
	}

	return userID, nil
}
