package mappers

import (
	"fmt"
	"strings"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
)

func RegisterClientRequestToAuth(request *gatewayv1.ClientRegisterRequest) *authv1.RegisterRequest {
	return &authv1.RegisterRequest{
		Email:    request.GetEmail(),
		Username: request.GetUsername(),
		Password: request.GetPassword(),
		Role:     authv1.UserRole_USER_ROLE_CLIENT,
	}
}

func RegisterTrainerRequestToAuth(request *gatewayv1.TrainerRegisterRequest) *authv1.RegisterRequest {
	return &authv1.RegisterRequest{
		Email:    request.GetEmail(),
		Username: request.GetUsername(),
		Password: request.GetPassword(),
		Role:     authv1.UserRole_USER_ROLE_TRAINER,
	}
}

func LoginRequestToAuth(request *gatewayv1.LoginRequest) *authv1.LoginRequest {
	return &authv1.LoginRequest{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	}
}

func AuthResponseFromServices(user *authv1.AuthUser, profile *profilev1.Profile) (*gatewayv1.AuthResponse, error) {
	mappedUser, err := UserFromServices(user, profile)
	if err != nil {
		return nil, err
	}

	return &gatewayv1.AuthResponse{User: mappedUser}, nil
}

func UserFromServices(user *authv1.AuthUser, profile *profilev1.Profile) (*gatewayv1.User, error) {
	if user == nil || profile == nil {
		return nil, fmt.Errorf("user and profile are required")
	}

	userID, err := int64ToInt32("auth.user_id", user.GetUserId())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.User{
		UserId:    userID,
		Username:  profile.GetUsername(),
		Email:     user.GetEmail(),
		CreatedAt: profile.GetCreatedAt(),
		UpdatedAt: profile.GetUpdatedAt(),
		IsTrainer: profile.GetIsTrainer(),
		IsAdmin:   user.GetRole() == authv1.UserRole_USER_ROLE_ADMIN,
		FirstName: profile.GetFirstName(),
		LastName:  profile.GetLastName(),
		Bio:       profile.Bio,
		AvatarUrl: profile.AvatarUrl,
	}, nil
}

func PasswordsMatch(password, passwordRepeat string) bool {
	return password == passwordRepeat
}

func RequireTrainerRole(user *authv1.AuthUser) error {
	if user == nil {
		return fmt.Errorf("user is required")
	}

	switch user.GetRole() {
	case authv1.UserRole_USER_ROLE_TRAINER, authv1.UserRole_USER_ROLE_ADMIN:
		return nil
	default:
		return fmt.Errorf("only trainer can perform this action")
	}
}

func NormalizeStatusCode(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "", "bad_request":
		return "bad_request"
	default:
		return code
	}
}
