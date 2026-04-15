package usecase

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUseCaseGetByIDMapsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockuserRepository(ctrl)
	repository.EXPECT().GetByID(gomock.Any(), int64(42)).Return(domain.User{}, sql.ErrNoRows)

	useCase := NewUserUseCase(repository, nil)

	_, err := useCase.GetByID(context.Background(), 42)
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("unexpected error: got %v, expect %v", err, ErrUserNotFound)
	}
}

func TestUserUseCaseRegisterClientSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockuserRepository(ctrl)

	repository.EXPECT().
		CreateClient(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, command CreateClientCommand) (int64, error) {
			if command.Username != "john_doe" || command.Email != "john@example.com" {
				t.Fatalf("unexpected command: %+v", command)
			}
			if err := bcrypt.CompareHashAndPassword([]byte(command.PasswordHash), []byte("supersecret123")); err != nil {
				t.Fatalf("password hash mismatch: %v", err)
			}

			return 7, nil
		})

	expectedUser := domain.User{ID: 7, Username: "john_doe", Email: "john@example.com"}
	repository.EXPECT().GetByID(gomock.Any(), int64(7)).Return(expectedUser, nil)

	useCase := NewUserUseCase(repository, nil)

	user, err := useCase.RegisterClient(context.Background(), RegisterClientCommand{
		Username:  "john_doe",
		Email:     "john@example.com",
		Password:  "supersecret123",
		FirstName: "John",
		LastName:  "Doe",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user != expectedUser {
		t.Fatalf("unexpected user: got %+v, expect %+v", user, expectedUser)
	}
}

func TestUserUseCaseRegisterTrainerMapsSportTypeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockuserRepository(ctrl)

	repository.EXPECT().
		CreateTrainer(gomock.Any(), gomock.Any()).
		Return(int64(0), ErrSportTypeNotFound)

	useCase := NewUserUseCase(repository, nil)

	_, err := useCase.RegisterTrainer(context.Background(), RegisterTrainerCommand{
		Username:        "coach",
		Email:           "coach@example.com",
		Password:        "supersecret123",
		FirstName:       "Coach",
		LastName:        "One",
		CareerSinceDate: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		Sports: []RegisterTrainerSportCommand{{
			SportTypeID:     99,
			ExperienceYears: 2,
		}},
	})
	if !errors.Is(err, ErrSportTypeNotFound) {
		t.Fatalf("unexpected error: got %v, expect %v", err, ErrSportTypeNotFound)
	}
}

func TestUserUseCaseAuthenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockuserRepository(ctrl)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("supersecret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}

	expectedUser := domain.User{ID: 3, Email: "john@example.com", PasswordHash: string(passwordHash)}
	repository.EXPECT().GetByEmail(gomock.Any(), "john@example.com").Return(expectedUser, nil)

	useCase := NewUserUseCase(repository, nil)

	user, err := useCase.Authenticate(context.Background(), "john@example.com", "supersecret123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != expectedUser.ID {
		t.Fatalf("unexpected user: got %+v, expect %+v", user, expectedUser)
	}
}

func TestUserUseCaseAuthenticateInvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockuserRepository(ctrl)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}

	repository.EXPECT().GetByEmail(gomock.Any(), "john@example.com").Return(domain.User{
		ID:           3,
		Email:        "john@example.com",
		PasswordHash: string(passwordHash),
	}, nil)

	useCase := NewUserUseCase(repository, nil)

	_, err = useCase.Authenticate(context.Background(), "john@example.com", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unexpected error: got %v, expect %v", err, ErrInvalidCredentials)
	}
}

func TestUserUseCaseUpdateProfileMapsUsernameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := NewMockuserRepository(ctrl)

	repository.EXPECT().
		UpdateProfile(gomock.Any(), int64(5), UpdateProfileCommand{HasUsername: true, Username: "taken"}).
		Return(ErrUsernameExists)

	useCase := NewUserUseCase(repository, nil)

	_, err := useCase.UpdateProfile(context.Background(), 5, UpdateProfileCommand{
		HasUsername: true,
		Username:    "taken",
	})
	if !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("unexpected error: got %v, expect %v", err, ErrUsernameExists)
	}
}

func TestUserUseCaseUploadAvatar(t *testing.T) {
	t.Run("storage not configured", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := NewMockuserRepository(ctrl)
		useCase := NewUserUseCase(repository, nil)

		_, err := useCase.UploadAvatar(context.Background(), 1, "avatar.jpg", "image/jpeg", bytes.NewReader([]byte("img")), 3)
		if err == nil || err.Error() != "avatar storage is not configured" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("user not found on update avatar url", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := NewMockuserRepository(ctrl)
		storage := NewMockavatarStorage(ctrl)

		storage.EXPECT().
			UploadAvatar(gomock.Any(), int64(9), "avatar.jpg", "image/jpeg", gomock.AssignableToTypeOf((*bytes.Reader)(nil)), int64(3)).
			Return("http://storage/avatars/file.jpg", nil)
		repository.EXPECT().
			UpdateAvatarURL(gomock.Any(), int64(9), "http://storage/avatars/file.jpg").
			Return(sql.ErrNoRows)

		useCase := NewUserUseCase(repository, storage)

		_, err := useCase.UploadAvatar(context.Background(), 9, "avatar.jpg", "image/jpeg", bytes.NewReader([]byte("img")), 3)
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("unexpected error: got %v, expect %v", err, ErrUserNotFound)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := NewMockuserRepository(ctrl)
		storage := NewMockavatarStorage(ctrl)

		expectedUser := domain.User{
			ID:        9,
			Username:  "john",
			AvatarURL: stringPtr("http://storage/avatars/file.jpg"),
		}

		storage.EXPECT().
			UploadAvatar(gomock.Any(), int64(9), "avatar.jpg", "image/jpeg", gomock.Any(), int64(3)).
			DoAndReturn(func(_ context.Context, _ int64, _ string, _ string, file io.Reader, _ int64) (string, error) {
				content, err := io.ReadAll(file)
				if err != nil {
					return "", err
				}
				if string(content) != "img" {
					t.Fatalf("unexpected file content: %q", string(content))
				}

				return "http://storage/avatars/file.jpg", nil
			})
		repository.EXPECT().
			UpdateAvatarURL(gomock.Any(), int64(9), "http://storage/avatars/file.jpg").
			Return(nil)
		repository.EXPECT().GetByID(gomock.Any(), int64(9)).Return(expectedUser, nil)

		useCase := NewUserUseCase(repository, storage)

		user, err := useCase.UploadAvatar(context.Background(), 9, "avatar.jpg", "image/jpeg", bytes.NewReader([]byte("img")), 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != expectedUser.ID || user.AvatarURL == nil || *user.AvatarURL != *expectedUser.AvatarURL {
			t.Fatalf("unexpected user: got %+v, expect %+v", user, expectedUser)
		}
	})
}

func stringPtr(value string) *string {
	return &value
}
