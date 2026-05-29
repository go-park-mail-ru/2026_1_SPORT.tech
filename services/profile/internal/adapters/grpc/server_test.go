package grpc_test

import (
	"context"
	"testing"
	"time"

	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/mocks"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerGetProfile(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	ctrl := gomock.NewController(t)
	profileUseCase := mocks.NewMockProfileUseCase(ctrl)
	profileUseCase.EXPECT().
		GetProfile(gomock.Any(), int64(7)).
		Return(domain.Profile{
			UserID:    7,
			Username:  "coach_john",
			FirstName: "John",
			LastName:  "Doe",
			IsTrainer: true,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil)

	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Profiles: profileUseCase,
		Authors:  profileUseCase,
		Avatars:  profileUseCase,
		Sports:   profileUseCase,
	})

	response, err := server.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetProfile().GetUserId() != 7 {
		t.Fatalf("unexpected user id: %d", response.GetProfile().GetUserId())
	}
}

func TestServerGetProfileMapsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	profileUseCase := mocks.NewMockProfileUseCase(ctrl)
	profileUseCase.EXPECT().
		GetProfile(gomock.Any(), int64(7)).
		Return(domain.Profile{}, domain.ErrProfileNotFound)

	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Profiles: profileUseCase,
		Authors:  profileUseCase,
		Avatars:  profileUseCase,
		Sports:   profileUseCase,
	})

	_, err := server.GetProfile(context.Background(), &profilev1.GetProfileRequest{UserId: 7})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("unexpected status code: %s", status.Code(err))
	}
}

func TestServerGetProfileByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	profileUseCase := mocks.NewMockProfileUseCase(ctrl)
	profileUseCase.EXPECT().
		GetProfileByUsername(gomock.Any(), "coach_john").
		Return(domain.Profile{UserID: 7, Username: "coach_john", FirstName: "John", LastName: "Doe"}, nil)

	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Profiles: profileUseCase,
		Authors:  profileUseCase,
		Avatars:  profileUseCase,
		Sports:   profileUseCase,
	})

	response, err := server.GetProfileByUsername(context.Background(), &profilev1.GetProfileByUsernameRequest{Username: "coach_john"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetProfile().GetUsername() != "coach_john" {
		t.Fatalf("unexpected username: %s", response.GetProfile().GetUsername())
	}
}
