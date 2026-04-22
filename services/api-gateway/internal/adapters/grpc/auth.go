package grpc

import (
	"context"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) GetCSRFToken(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.CSRFTokenResponse, error) {
	csrfToken, err := issueCSRFToken(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &gatewayv1.CSRFTokenResponse{CsrfToken: csrfToken}, nil
}

func (server *Server) RegisterClient(ctx context.Context, request *gatewayv1.ClientRegisterRequest) (*gatewayv1.AuthResponse, error) {
	if !mappers.PasswordsMatch(request.GetPassword(), request.GetPasswordRepeat()) {
		return nil, status.Error(codes.InvalidArgument, "passwords do not match")
	}

	authResponse, err := server.authClient.Register(forwardContext(ctx), mappers.RegisterClientRequestToAuth(request))
	if err != nil {
		return nil, err
	}

	profileRequest, err := mappers.CreateProfileRequestToProfile(
		authResponse.GetUser().GetUserId(),
		request.GetUsername(),
		request.GetFirstName(),
		request.GetLastName(),
		false,
		nil,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	profileResponse, err := server.profileClient.CreateProfile(forwardContext(ctx), profileRequest)
	if err != nil {
		return nil, err
	}

	if err := setSessionCookie(ctx, authResponse.GetSession().GetSessionToken(), authResponse.GetSession().GetExpiresAt()); err != nil {
		return nil, status.Errorf(codes.Internal, "set session cookie: %v", err)
	}
	if err := setHTTPStatus(ctx, 201); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return mappers.AuthResponseFromServices(authResponse.GetUser(), profileResponse.GetProfile())
}

func (server *Server) RegisterTrainer(ctx context.Context, request *gatewayv1.TrainerRegisterRequest) (*gatewayv1.AuthResponse, error) {
	if !mappers.PasswordsMatch(request.GetPassword(), request.GetPasswordRepeat()) {
		return nil, status.Error(codes.InvalidArgument, "passwords do not match")
	}

	authResponse, err := server.authClient.Register(forwardContext(ctx), mappers.RegisterTrainerRequestToAuth(request))
	if err != nil {
		return nil, err
	}

	profileRequest, err := mappers.CreateProfileRequestToProfile(
		authResponse.GetUser().GetUserId(),
		request.GetUsername(),
		request.GetFirstName(),
		request.GetLastName(),
		true,
		request.GetTrainerDetails(),
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	profileResponse, err := server.profileClient.CreateProfile(forwardContext(ctx), profileRequest)
	if err != nil {
		return nil, err
	}

	if err := setSessionCookie(ctx, authResponse.GetSession().GetSessionToken(), authResponse.GetSession().GetExpiresAt()); err != nil {
		return nil, status.Errorf(codes.Internal, "set session cookie: %v", err)
	}
	if err := setHTTPStatus(ctx, 201); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return mappers.AuthResponseFromServices(authResponse.GetUser(), profileResponse.GetProfile())
}

func (server *Server) Login(ctx context.Context, request *gatewayv1.LoginRequest) (*gatewayv1.AuthResponse, error) {
	authResponse, err := server.authClient.Login(forwardContext(ctx), mappers.LoginRequestToAuth(request))
	if err != nil {
		return nil, err
	}

	profile, err := server.getProfile(ctx, authResponse.GetUser().GetUserId())
	if err != nil {
		return nil, err
	}

	if err := setSessionCookie(ctx, authResponse.GetSession().GetSessionToken(), authResponse.GetSession().GetExpiresAt()); err != nil {
		return nil, status.Errorf(codes.Internal, "set session cookie: %v", err)
	}

	return mappers.AuthResponseFromServices(authResponse.GetUser(), profile)
}

func (server *Server) GetMe(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.AuthResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := server.getProfile(ctx, principal.User.GetUserId())
	if err != nil {
		return nil, err
	}

	return mappers.AuthResponseFromServices(principal.User, profile)
}

func (server *Server) Logout(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := server.authClient.Logout(
		forwardContext(ctx),
		&authv1.LogoutRequest{SessionToken: principal.SessionToken},
	); err != nil {
		return nil, err
	}

	if err := clearSessionCookie(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "clear session cookie: %v", err)
	}
	if err := clearCSRFCookie(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "clear csrf cookie: %v", err)
	}
	if err := setHTTPStatus(ctx, 204); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return &emptypb.Empty{}, nil
}
