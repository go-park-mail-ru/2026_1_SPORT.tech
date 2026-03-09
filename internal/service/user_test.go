package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
)

type userRepositoryStub struct {
	getByIDFunc func(ctx context.Context, userID int64) (repository.User, error)
}

func (stub *userRepositoryStub) GetByID(ctx context.Context, userID int64) (repository.User, error) {
	if stub.getByIDFunc == nil {
		return repository.User{}, nil
	}

	return stub.getByIDFunc(ctx, userID)
}

type getUserByIDTest struct {
	name       string
	userID     int64
	repository *userRepositoryStub
	expect     User
	expectErr  error
}

func TestUserServiceGetByIDPositive(t *testing.T) {
	bio := "bio"
	avatarURL := "https://example.com/avatar.jpg"
	createdAt := time.Date(2026, time.March, 9, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.March, 9, 13, 0, 0, 0, time.UTC)

	tests := []getUserByIDTest{
		{
			name:   "Корректный маппинг пользователя",
			userID: 42,
			repository: &userRepositoryStub{
				getByIDFunc: func(ctx context.Context, userID int64) (repository.User, error) {
					return repository.User{
						ID:        42,
						Username:  "john_doe",
						Email:     "user@example.com",
						CreatedAt: createdAt,
						UpdatedAt: updatedAt,
						IsTrainer: true,
						IsAdmin:   false,
						Profile: repository.UserProfile{
							Username:  "john_doe",
							FirstName: "John",
							LastName:  "Doe",
							Bio:       &bio,
							AvatarURL: &avatarURL,
						},
					}, nil
				},
			},
			expect: User{
				ID:        42,
				Username:  "john_doe",
				Email:     "user@example.com",
				CreatedAt: createdAt,
				UpdatedAt: updatedAt,
				IsTrainer: true,
				IsAdmin:   false,
				Profile: UserProfile{
					Username:  "john_doe",
					FirstName: "John",
					LastName:  "Doe",
					Bio:       &bio,
					AvatarURL: &avatarURL,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUserService(tt.repository)

			res, err := service.GetByID(context.Background(), tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
			if res != tt.expect {
				t.Fatalf("unexpected result: got %+v, expect %+v", res, tt.expect)
			}
		})
	}
}

func TestUserServiceGetByIDNegative(t *testing.T) {
	expectedErr := errors.New("get user")
	tests := []getUserByIDTest{
		{
			name:   "Пользователь не найден",
			userID: 42,
			repository: &userRepositoryStub{
				getByIDFunc: func(ctx context.Context, userID int64) (repository.User, error) {
					return repository.User{}, sql.ErrNoRows
				},
			},
			expectErr: ErrUserNotFound,
		},
		{
			name:   "Ошибка репозитория",
			userID: 42,
			repository: &userRepositoryStub{
				getByIDFunc: func(ctx context.Context, userID int64) (repository.User, error) {
					return repository.User{}, expectedErr
				},
			},
			expectErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewUserService(tt.repository)

			_, err := service.GetByID(context.Background(), tt.userID)
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}
