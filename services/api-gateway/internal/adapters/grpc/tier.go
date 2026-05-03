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

func (server *Server) ListTiers(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.TiersResponse, error) {
	trainerUserID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListSubscriptionTiers(
		forwardContext(ctx),
		&contentv1.ListSubscriptionTiersRequest{TrainerUserId: trainerUserID},
	)
	if err != nil {
		return nil, err
	}

	return mappers.TiersResponseFromContent(response)
}

func (server *Server) CreateTier(ctx context.Context, request *gatewayv1.CreateTierRequest) (*gatewayv1.Tier, error) {
	trainerUserID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.CreateSubscriptionTier(
		forwardContext(ctx),
		mappers.CreateTierRequestToContent(trainerUserID, request),
	)
	if err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 201); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return mappers.TierFromContent(response)
}

func (server *Server) UpdateTier(ctx context.Context, request *gatewayv1.UpdateTierRequest) (*gatewayv1.Tier, error) {
	trainerUserID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.UpdateSubscriptionTier(
		forwardContext(ctx),
		mappers.UpdateTierRequestToContent(trainerUserID, request),
	)
	if err != nil {
		return nil, err
	}

	return mappers.TierFromContent(response)
}

func (server *Server) DeleteTier(ctx context.Context, request *gatewayv1.DeleteTierRequest) (*emptypb.Empty, error) {
	trainerUserID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := server.contentClient.DeleteSubscriptionTier(
		forwardContext(ctx),
		&contentv1.DeleteSubscriptionTierRequest{
			TrainerUserId: trainerUserID,
			TierId:        int64(request.GetTierId()),
		},
	); err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 204); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) requireTrainerUserID(ctx context.Context) (int64, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return 0, err
	}
	if err := mappers.RequireTrainerRole(principal.User); err != nil {
		return 0, status.Error(codes.PermissionDenied, err.Error())
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "unauthorized")
	}

	return userID, nil
}
