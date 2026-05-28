package grpc

import (
	"context"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RegistrationUseCase interface {
	Register(ctx context.Context, command usecase.RegisterCommand) (usecase.AuthResult, error)
}

type LoginUseCase interface {
	Login(ctx context.Context, command usecase.LoginCommand) (usecase.AuthResult, error)
}

type SessionUseCase interface {
	Logout(ctx context.Context, command usecase.LogoutCommand) error
	GetSession(ctx context.Context, query usecase.GetSessionQuery) (usecase.SessionResult, error)
}

type AccountUseCase interface {
	ChangePassword(ctx context.Context, command usecase.ChangePasswordCommand) error
	ChangeEmail(ctx context.Context, command usecase.ChangeEmailCommand) (domain.Account, error)
	PromoteToTrainer(ctx context.Context, command usecase.PromoteToTrainerCommand) (domain.Account, error)
	LogoutAll(ctx context.Context, command usecase.LogoutAllCommand) error
	DeleteAccount(ctx context.Context, command usecase.DeleteAccountCommand) error
}

type UseCases struct {
	Registration RegistrationUseCase
	Login        LoginUseCase
	Session      SessionUseCase
	Account      AccountUseCase
}

type Server struct {
	authv1.UnimplementedAuthServiceServer
	useCases UseCases
}

func NewServer(useCases UseCases) *Server {
	return &Server{useCases: useCases}
}

func (server *Server) Register(ctx context.Context, request *authv1.RegisterRequest) (*authv1.AuthSessionResponse, error) {
	command, err := mappers.RegisterRequestToCommand(request)
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	result, err := server.useCases.Registration.Register(ctx, command)
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewAuthSessionResponse(result), nil
}

func (server *Server) Login(ctx context.Context, request *authv1.LoginRequest) (*authv1.AuthSessionResponse, error) {
	result, err := server.useCases.Login.Login(ctx, mappers.LoginRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewAuthSessionResponse(result), nil
}

func (server *Server) Logout(ctx context.Context, request *authv1.LogoutRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Session.Logout(ctx, mappers.LogoutRequestToCommand(request)); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) GetSession(ctx context.Context, request *authv1.GetSessionRequest) (*authv1.GetSessionResponse, error) {
	result, err := server.useCases.Session.GetSession(ctx, mappers.GetSessionRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewGetSessionResponse(result), nil
}

func (server *Server) ChangePassword(ctx context.Context, request *authv1.ChangePasswordRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Account.ChangePassword(ctx, usecase.ChangePasswordCommand{
		UserID:          request.GetUserId(),
		CurrentPassword: request.GetCurrentPassword(),
		NewPassword:     request.GetNewPassword(),
	}); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) ChangeEmail(ctx context.Context, request *authv1.ChangeEmailRequest) (*authv1.AuthUser, error) {
	account, err := server.useCases.Account.ChangeEmail(ctx, usecase.ChangeEmailCommand{
		UserID:          request.GetUserId(),
		CurrentPassword: request.GetCurrentPassword(),
		NewEmail:        request.GetNewEmail(),
	})
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewAuthUser(account), nil
}

func (server *Server) PromoteToTrainer(ctx context.Context, request *authv1.PromoteToTrainerRequest) (*authv1.AuthUser, error) {
	account, err := server.useCases.Account.PromoteToTrainer(ctx, usecase.PromoteToTrainerCommand{
		UserID: request.GetUserId(),
	})
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewAuthUser(account), nil
}

func (server *Server) LogoutAll(ctx context.Context, request *authv1.LogoutAllRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Account.LogoutAll(ctx, usecase.LogoutAllCommand{
		UserID: request.GetUserId(),
	}); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) DeleteAccount(ctx context.Context, request *authv1.DeleteAccountRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Account.DeleteAccount(ctx, usecase.DeleteAccountCommand{
		UserID:          request.GetUserId(),
		CurrentPassword: request.GetCurrentPassword(),
	}); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return &emptypb.Empty{}, nil
}
