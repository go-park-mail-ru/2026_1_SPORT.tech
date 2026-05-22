package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) GetMyStatistics(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.StatisticsResponse, error) {
	trainerUserID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.GetTrainerStatistics(
		forwardContext(ctx),
		&contentv1.GetTrainerStatisticsRequest{TrainerUserId: trainerUserID, Currency: "RUB"},
	)
	if err != nil {
		return nil, err
	}

	return mappers.StatisticsResponseFromContent(response)
}
