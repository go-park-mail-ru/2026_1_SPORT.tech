package grpc

import (
	"context"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) DonateToProfile(context.Context, *gatewayv1.DonateToProfileRequest) (*gatewayv1.DonationResponse, error) {
	return nil, status.Error(codes.Unimplemented, "billing service is not implemented yet")
}
