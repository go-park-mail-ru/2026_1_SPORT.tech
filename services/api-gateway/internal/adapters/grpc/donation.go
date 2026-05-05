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

func (server *Server) DonateToProfile(ctx context.Context, request *gatewayv1.DonateToProfileRequest) (*gatewayv1.DonationResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.contentClient.DonateToProfile(
		forwardContext(ctx),
		mappers.DonateToProfileRequestToContent(userID, request),
	)
	if err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 201); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return mappers.DonationResponseFromContent(response)
}

func (server *Server) GetMyBalance(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.BalanceResponse, error) {
	trainerUserID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.GetBalance(
		forwardContext(ctx),
		&contentv1.GetBalanceRequest{TrainerUserId: trainerUserID, Currency: "RUB"},
	)
	if err != nil {
		return nil, err
	}

	return mappers.BalanceResponseFromContent(response)
}
