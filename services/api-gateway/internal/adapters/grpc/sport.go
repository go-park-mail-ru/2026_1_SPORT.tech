package grpc

import (
	"context"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) ListSportTypes(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.SportTypesResponse, error) {
	response, err := server.profileClient.ListSportTypes(forwardContext(ctx), &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return mappers.SportTypesResponseFromProfile(response)
}
