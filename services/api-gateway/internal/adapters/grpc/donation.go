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
	if _, err := server.requireSession(ctx); err != nil {
		return nil, err
	}

	return nil, status.Error(codes.FailedPrecondition, "create and confirm a donation payment instead")
}

func (server *Server) ListMyReceivedDonations(ctx context.Context, request *gatewayv1.ListDonationsRequest) (*gatewayv1.ListDonationsResponse, error) {
	trainerUserID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListReceivedDonations(
		forwardContext(ctx),
		&contentv1.ListReceivedDonationsRequest{
			TrainerUserId: trainerUserID,
			Limit:         request.GetLimit(),
			Offset:        request.GetOffset(),
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.ListDonationsResponseFromContent(response)
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
