package mappers

import (
	"errors"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func RegisterRequestToCommand(request *authv1.RegisterRequest) (usecase.RegisterCommand, error) {
	role, err := roleFromProto(request.GetRole())
	if err != nil {
		return usecase.RegisterCommand{}, err
	}

	return usecase.RegisterCommand{
		Email:    request.GetEmail(),
		Username: request.GetUsername(),
		Password: request.GetPassword(),
		Role:     role,
	}, nil
}

func LoginRequestToCommand(request *authv1.LoginRequest) usecase.LoginCommand {
	return usecase.LoginCommand{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	}
}

func LogoutRequestToCommand(request *authv1.LogoutRequest) usecase.LogoutCommand {
	return usecase.LogoutCommand{
		SessionToken: request.GetSessionToken(),
	}
}

func GetSessionRequestToQuery(request *authv1.GetSessionRequest) usecase.GetSessionQuery {
	return usecase.GetSessionQuery{
		SessionToken: request.GetSessionToken(),
	}
}

func NewAuthSessionResponse(result usecase.AuthResult) *authv1.AuthSessionResponse {
	return &authv1.AuthSessionResponse{
		User: buildAuthUser(result.Account),
		Session: &authv1.SessionInfo{
			SessionToken: result.SessionToken,
			ExpiresAt:    timestamppb.New(result.SessionExpiresAt),
		},
	}
}

func NewGetSessionResponse(result usecase.SessionResult) *authv1.GetSessionResponse {
	return &authv1.GetSessionResponse{
		User: buildAuthUser(result.Account),
		Session: &authv1.SessionInfo{
			ExpiresAt: timestamppb.New(result.Session.ExpiresAt),
		},
	}
}

func ErrorToStatus(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, usecase.ErrInvalidEmail),
		errors.Is(err, usecase.ErrInvalidUsername),
		errors.Is(err, usecase.ErrWeakPassword),
		errors.Is(err, usecase.ErrMissingSessionToken),
		errors.Is(err, domain.ErrInvalidRole):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrEmailTaken), errors.Is(err, domain.ErrUsernameTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials),
		errors.Is(err, domain.ErrSessionNotFound),
		errors.Is(err, domain.ErrSessionExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrAccountNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrAccountDisabled):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func buildAuthUser(account domain.Account) *authv1.AuthUser {
	return &authv1.AuthUser{
		UserId:   account.ID,
		Email:    account.Email,
		Username: account.Username,
		Role:     roleToProto(account.Role),
		Status:   statusToProto(account.Status),
	}
}

func roleFromProto(role authv1.UserRole) (domain.Role, error) {
	switch role {
	case authv1.UserRole_USER_ROLE_UNSPECIFIED, authv1.UserRole_USER_ROLE_CLIENT:
		return domain.RoleClient, nil
	case authv1.UserRole_USER_ROLE_TRAINER:
		return domain.RoleTrainer, nil
	case authv1.UserRole_USER_ROLE_ADMIN:
		return domain.RoleAdmin, nil
	default:
		return "", domain.ErrInvalidRole
	}
}

func roleToProto(role domain.Role) authv1.UserRole {
	switch role {
	case domain.RoleTrainer:
		return authv1.UserRole_USER_ROLE_TRAINER
	case domain.RoleAdmin:
		return authv1.UserRole_USER_ROLE_ADMIN
	default:
		return authv1.UserRole_USER_ROLE_CLIENT
	}
}

func statusToProto(accountStatus domain.Status) authv1.AccountStatus {
	switch accountStatus {
	case domain.StatusDisabled:
		return authv1.AccountStatus_ACCOUNT_STATUS_DISABLED
	default:
		return authv1.AccountStatus_ACCOUNT_STATUS_ACTIVE
	}
}
