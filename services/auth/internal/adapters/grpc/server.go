package grpc

import (
	"context"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthUseCase interface {
	Register(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error)
	Login(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error)
	Logout(ctx context.Context, command usecase.LogoutCommand) error
	GetSession(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error)
}

type Server struct {
	authv1.UnimplementedAuthServiceServer
	authUseCase AuthUseCase
}

func NewServer(authUseCase AuthUseCase) *Server {
	return &Server{authUseCase: authUseCase}
}

func (server *Server) Register(ctx context.Context, request *authv1.RegisterRequest) (*authv1.AuthSessionResponse, error) {
	command, err := mappers.RegisterRequestToCommand(request)
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	result, err := server.authUseCase.Register(ctx, command)
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewAuthSessionResponse(result), nil
}

func (server *Server) Login(ctx context.Context, request *authv1.LoginRequest) (*authv1.AuthSessionResponse, error) {
	result, err := server.authUseCase.Login(ctx, mappers.LoginRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewAuthSessionResponse(result), nil
}

func (server *Server) Logout(ctx context.Context, request *authv1.LogoutRequest) (*emptypb.Empty, error) {
	if err := server.authUseCase.Logout(ctx, mappers.LogoutRequestToCommand(request)); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) GetSession(ctx context.Context, request *authv1.GetSessionRequest) (*authv1.GetSessionResponse, error) {
	result, err := server.authUseCase.GetSession(ctx, mappers.GetSessionRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewGetSessionResponse(result), nil
}
