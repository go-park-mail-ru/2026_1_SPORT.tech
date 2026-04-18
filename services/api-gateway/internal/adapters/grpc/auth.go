package grpc

import (
	"context"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) Register(ctx context.Context, request *gatewayv1.RegisterRequest) (*gatewayv1.AuthSessionResponse, error) {
	registerRequest, err := mappers.RegisterRequestToAuth(request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := server.authClient.Register(forwardContext(ctx), registerRequest)
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.AuthSessionResponseFromAuth(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map auth register response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) Login(ctx context.Context, request *gatewayv1.LoginRequest) (*gatewayv1.AuthSessionResponse, error) {
	response, err := server.authClient.Login(forwardContext(ctx), mappers.LoginRequestToAuth(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.AuthSessionResponseFromAuth(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map auth login response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) Logout(ctx context.Context, request *gatewayv1.LogoutRequest) (*emptypb.Empty, error) {
	return server.authClient.Logout(forwardContext(ctx), mappers.LogoutRequestToAuth(request))
}

func (server *Server) ResolveSession(ctx context.Context, request *gatewayv1.ResolveSessionRequest) (*gatewayv1.ResolveSessionResponse, error) {
	response, err := server.authClient.GetSession(forwardContext(ctx), mappers.ResolveSessionRequestToAuth(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.ResolveSessionResponseFromAuth(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map auth resolve session response: %v", err)
	}

	return mappedResponse, nil
}
