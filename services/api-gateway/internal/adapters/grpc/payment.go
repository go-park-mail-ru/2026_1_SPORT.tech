package grpc

import (
	"context"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateDonationPayment(ctx context.Context, request *gatewayv1.CreateDonationPaymentRequest) (*gatewayv1.PaymentResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.contentClient.CreateDonationPayment(
		forwardContext(ctx),
		mappers.CreateDonationPaymentRequestToContent(userID, request),
	)
	if err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 201); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return mappers.PaymentResponseFromContent(response)
}

func (server *Server) ConfirmDonationPayment(ctx context.Context, request *gatewayv1.ConfirmDonationPaymentRequest) (*gatewayv1.PaymentResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.contentClient.ConfirmDonationPayment(
		forwardContext(ctx),
		mappers.ConfirmDonationPaymentRequestToContent(userID, request),
	)
	if err != nil {
		return nil, err
	}

	return mappers.PaymentResponseFromContent(response)
}
