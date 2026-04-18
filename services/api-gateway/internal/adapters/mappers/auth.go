package mappers

import (
	"fmt"
	"strings"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func RegisterRequestToAuth(request *gatewayv1.RegisterRequest) (*authv1.RegisterRequest, error) {
	role, err := publicRoleToAuthRole(request.GetRole())
	if err != nil {
		return nil, err
	}

	return &authv1.RegisterRequest{
		Email:    request.GetEmail(),
		Username: request.GetUsername(),
		Password: request.GetPassword(),
		Role:     role,
	}, nil
}

func LoginRequestToAuth(request *gatewayv1.LoginRequest) *authv1.LoginRequest {
	return &authv1.LoginRequest{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	}
}

func LogoutRequestToAuth(request *gatewayv1.LogoutRequest) *authv1.LogoutRequest {
	return &authv1.LogoutRequest{SessionToken: request.GetSessionToken()}
}

func ResolveSessionRequestToAuth(request *gatewayv1.ResolveSessionRequest) *authv1.GetSessionRequest {
	return &authv1.GetSessionRequest{SessionToken: request.GetSessionToken()}
}

func AuthSessionResponseFromAuth(response *authv1.AuthSessionResponse) (*gatewayv1.AuthSessionResponse, error) {
	if response == nil {
		return nil, nil
	}

	user, err := authUserFromAuth(response.GetUser())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.AuthSessionResponse{
		User:    user,
		Session: sessionInfoFromAuth(response.GetSession()),
	}, nil
}

func ResolveSessionResponseFromAuth(response *authv1.GetSessionResponse) (*gatewayv1.ResolveSessionResponse, error) {
	if response == nil {
		return nil, nil
	}

	user, err := authUserFromAuth(response.GetUser())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.ResolveSessionResponse{
		User:    user,
		Session: sessionInfoFromAuth(response.GetSession()),
	}, nil
}

func authUserFromAuth(user *authv1.AuthUser) (*gatewayv1.AuthUser, error) {
	if user == nil {
		return nil, nil
	}

	userID, err := int64ToInt32("auth.user_id", user.GetUserId())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.AuthUser{
		UserId:   userID,
		Email:    user.GetEmail(),
		Username: user.GetUsername(),
		Role:     authRoleToPublicRole(user.GetRole()),
		Status:   authStatusToPublicStatus(user.GetStatus()),
	}, nil
}

func sessionInfoFromAuth(session *authv1.SessionInfo) *gatewayv1.SessionInfo {
	if session == nil {
		return nil
	}

	return &gatewayv1.SessionInfo{
		SessionToken: session.GetSessionToken(),
		ExpiresAt:    session.GetExpiresAt(),
	}
}

func publicRoleToAuthRole(role string) (authv1.UserRole, error) {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "", "client", "user_role_client":
		return authv1.UserRole_USER_ROLE_CLIENT, nil
	case "trainer", "user_role_trainer":
		return authv1.UserRole_USER_ROLE_TRAINER, nil
	case "admin", "user_role_admin":
		return authv1.UserRole_USER_ROLE_ADMIN, nil
	default:
		return authv1.UserRole_USER_ROLE_UNSPECIFIED, fmt.Errorf("invalid role %q", role)
	}
}

func authRoleToPublicRole(role authv1.UserRole) string {
	switch role {
	case authv1.UserRole_USER_ROLE_TRAINER:
		return "trainer"
	case authv1.UserRole_USER_ROLE_ADMIN:
		return "admin"
	default:
		return "client"
	}
}

func authStatusToPublicStatus(status authv1.AccountStatus) string {
	switch status {
	case authv1.AccountStatus_ACCOUNT_STATUS_DISABLED:
		return "disabled"
	default:
		return "active"
	}
}
