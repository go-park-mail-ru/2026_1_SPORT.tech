package grpc

import (
	"context"
	"testing"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestSubscriptionLevelFromContext(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-subscription-level", "2"))

	level, err := subscriptionLevelFromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if level == nil || *level != 2 {
		t.Fatalf("unexpected subscription level: %v", level)
	}
}

func TestSubscriptionLevelFromContextRejectsInvalidValue(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-subscription-level", "bad"))

	_, err := subscriptionLevelFromContext(ctx)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unexpected status: %s", status.Code(err))
	}
}

func TestRequireSubscriptionUserIDAllowsTrainer(t *testing.T) {
	server := NewServer(stubAuthServiceClient{
		getSessionFunc: func(_ context.Context, request *authv1.GetSessionRequest, _ ...grpc.CallOption) (*authv1.GetSessionResponse, error) {
			if request.GetSessionToken() != "trainer-session" {
				t.Fatalf("unexpected session token: %s", request.GetSessionToken())
			}

			return &authv1.GetSessionResponse{
				User: &authv1.AuthUser{
					UserId: 10,
					Role:   authv1.UserRole_USER_ROLE_TRAINER,
					Status: authv1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
				},
				Session: &authv1.SessionInfo{SessionToken: request.GetSessionToken()},
			}, nil
		},
	}, nil, nil)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-session-token", "trainer-session"))

	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 10 {
		t.Fatalf("unexpected user id: %d", userID)
	}
}

func TestRegisterClientRejectsInvalidProfileBeforeAuthWrite(t *testing.T) {
	server := NewServer(stubAuthServiceClient{}, nil, nil)

	_, err := server.RegisterClient(context.Background(), &gatewayv1.ClientRegisterRequest{
		Username:       "valid_user",
		Email:          "client@example.com",
		Password:       "supersecret123",
		PasswordRepeat: "supersecret123",
		FirstName:      "",
		LastName:       "User",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unexpected status: %s", status.Code(err))
	}
}

func TestRegisterTrainerRejectsInvalidDetailsBeforeAuthWrite(t *testing.T) {
	server := NewServer(stubAuthServiceClient{}, nil, nil)
	futureDate := "2999-01-01"

	_, err := server.RegisterTrainer(context.Background(), &gatewayv1.TrainerRegisterRequest{
		Username:       "coach_valid",
		Email:          "trainer@example.com",
		Password:       "supersecret123",
		PasswordRepeat: "supersecret123",
		FirstName:      "Coach",
		LastName:       "Valid",
		TrainerDetails: &gatewayv1.TrainerDetails{
			CareerSinceDate: &futureDate,
		},
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unexpected status: %s", status.Code(err))
	}
}

type stubAuthServiceClient struct {
	getSessionFunc func(context.Context, *authv1.GetSessionRequest, ...grpc.CallOption) (*authv1.GetSessionResponse, error)
}

func (client stubAuthServiceClient) Register(context.Context, *authv1.RegisterRequest, ...grpc.CallOption) (*authv1.AuthSessionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "register is not implemented")
}

func (client stubAuthServiceClient) Login(context.Context, *authv1.LoginRequest, ...grpc.CallOption) (*authv1.AuthSessionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "login is not implemented")
}

func (client stubAuthServiceClient) Logout(context.Context, *authv1.LogoutRequest, ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "logout is not implemented")
}

func (client stubAuthServiceClient) GetSession(ctx context.Context, request *authv1.GetSessionRequest, opts ...grpc.CallOption) (*authv1.GetSessionResponse, error) {
	if client.getSessionFunc == nil {
		return nil, status.Error(codes.Unimplemented, "get session is not implemented")
	}

	return client.getSessionFunc(ctx, request, opts...)
}
