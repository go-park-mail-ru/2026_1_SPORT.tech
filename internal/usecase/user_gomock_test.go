package usecase_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/gen"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
	"github.com/golang/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUseCaseGetByIDMapsNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := gen.NewMockuserRepository(ctrl)
	repository.EXPECT().GetByID(gomock.Any(), int64(42)).Return(domain.User{}, sql.ErrNoRows)

	useCase := usecase.NewUserUseCase(repository, nil)

	_, err := useCase.GetByID(context.Background(), 42)
	if !errors.Is(err, usecase.ErrUserNotFound) {
		t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrUserNotFound)
	}
}

func TestUserUseCaseListTrainers(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := gen.NewMockuserRepository(ctrl)
	expected := []domain.TrainerListItem{{ID: 7, Username: "coach"}}
	repository.EXPECT().ListTrainers(gomock.Any()).Return(expected, nil)

	useCase := usecase.NewUserUseCase(repository, nil)

	trainers, err := useCase.ListTrainers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(trainers) != 1 || trainers[0].ID != 7 {
		t.Fatalf("unexpected trainers: %+v", trainers)
	}
}

func TestUserUseCaseRegisterClientSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := gen.NewMockuserRepository(ctrl)

	repository.EXPECT().
		CreateClient(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, command usecase.CreateClientCommand) (int64, error) {
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

	useCase := usecase.NewUserUseCase(repository, nil)

	user, err := useCase.RegisterClient(context.Background(), usecase.RegisterClientCommand{
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
	repository := gen.NewMockuserRepository(ctrl)

	repository.EXPECT().
		CreateTrainer(gomock.Any(), gomock.Any()).
		Return(int64(0), usecase.ErrSportTypeNotFound)

	useCase := usecase.NewUserUseCase(repository, nil)

	_, err := useCase.RegisterTrainer(context.Background(), usecase.RegisterTrainerCommand{
		Username:        "coach",
		Email:           "coach@example.com",
		Password:        "supersecret123",
		FirstName:       "Coach",
		LastName:        "One",
		CareerSinceDate: time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		Sports: []usecase.RegisterTrainerSportCommand{{
			SportTypeID:     99,
			ExperienceYears: 2,
		}},
	})
	if !errors.Is(err, usecase.ErrSportTypeNotFound) {
		t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrSportTypeNotFound)
	}
}

func TestUserUseCaseAuthenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := gen.NewMockuserRepository(ctrl)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("supersecret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}

	expectedUser := domain.User{ID: 3, Email: "john@example.com", PasswordHash: string(passwordHash)}
	repository.EXPECT().GetByEmail(gomock.Any(), "john@example.com").Return(expectedUser, nil)

	useCase := usecase.NewUserUseCase(repository, nil)

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
	repository := gen.NewMockuserRepository(ctrl)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}

	repository.EXPECT().GetByEmail(gomock.Any(), "john@example.com").Return(domain.User{
		ID:           3,
		Email:        "john@example.com",
		PasswordHash: string(passwordHash),
	}, nil)

	useCase := usecase.NewUserUseCase(repository, nil)

	_, err = useCase.Authenticate(context.Background(), "john@example.com", "wrong-password")
	if !errors.Is(err, usecase.ErrInvalidCredentials) {
		t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrInvalidCredentials)
	}
}

func TestUserUseCaseUpdateProfileMapsUsernameExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := gen.NewMockuserRepository(ctrl)

	repository.EXPECT().
		UpdateProfile(gomock.Any(), int64(5), usecase.UpdateProfileCommand{HasUsername: true, Username: "taken"}).
		Return(usecase.ErrUsernameExists)

	useCase := usecase.NewUserUseCase(repository, nil)

	_, err := useCase.UpdateProfile(context.Background(), 5, usecase.UpdateProfileCommand{
		HasUsername: true,
		Username:    "taken",
	})
	if !errors.Is(err, usecase.ErrUsernameExists) {
		t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrUsernameExists)
	}
}

func TestUserUseCaseUpdateProfileTrainerFieldsRequireTrainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := gen.NewMockuserRepository(ctrl)
	repository.EXPECT().GetByID(gomock.Any(), int64(5)).Return(domain.User{ID: 5, IsTrainer: false}, nil)

	useCase := usecase.NewUserUseCase(repository, nil)

	_, err := useCase.UpdateProfile(context.Background(), 5, usecase.UpdateProfileCommand{
		HasCareerSinceDate: true,
		CareerSinceDate:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, usecase.ErrTrainerProfileForbidden) {
		t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrTrainerProfileForbidden)
	}
}

func TestUserUseCaseUploadAvatar(t *testing.T) {
	t.Run("storage not configured", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockuserRepository(ctrl)
		useCase := usecase.NewUserUseCase(repository, nil)

		_, err := useCase.UploadAvatar(context.Background(), 1, "avatar.jpg", "image/jpeg", bytes.NewReader([]byte("img")), 3)
		if err == nil || err.Error() != "avatar storage is not configured" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("user not found on update avatar url", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockuserRepository(ctrl)
		storage := gen.NewMockavatarStorage(ctrl)

		storage.EXPECT().
			UploadAvatar(gomock.Any(), int64(9), "avatar.jpg", "image/jpeg", gomock.AssignableToTypeOf((*bytes.Reader)(nil)), int64(3)).
			Return("http://storage/avatars/file.jpg", nil)
		repository.EXPECT().
			UpdateAvatarURL(gomock.Any(), int64(9), "http://storage/avatars/file.jpg").
			Return(sql.ErrNoRows)

		useCase := usecase.NewUserUseCase(repository, storage)

		_, err := useCase.UploadAvatar(context.Background(), 9, "avatar.jpg", "image/jpeg", bytes.NewReader([]byte("img")), 3)
		if !errors.Is(err, usecase.ErrUserNotFound) {
			t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrUserNotFound)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockuserRepository(ctrl)
		storage := gen.NewMockavatarStorage(ctrl)

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

		useCase := usecase.NewUserUseCase(repository, storage)

		user, err := useCase.UploadAvatar(context.Background(), 9, "avatar.jpg", "image/jpeg", bytes.NewReader([]byte("img")), 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.ID != expectedUser.ID || user.AvatarURL == nil || *user.AvatarURL != *expectedUser.AvatarURL {
			t.Fatalf("unexpected user: got %+v, expect %+v", user, expectedUser)
		}
	})
}

func TestUserUseCaseDeleteAvatar(t *testing.T) {
	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockuserRepository(ctrl)
		storage := gen.NewMockavatarStorage(ctrl)

		repository.EXPECT().GetByID(gomock.Any(), int64(9)).Return(domain.User{}, sql.ErrNoRows)

		useCase := usecase.NewUserUseCase(repository, storage)

		err := useCase.DeleteAvatar(context.Background(), 9)
		if !errors.Is(err, usecase.ErrUserNotFound) {
			t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrUserNotFound)
		}
	})

	t.Run("no avatar is no-op", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockuserRepository(ctrl)
		storage := gen.NewMockavatarStorage(ctrl)

		repository.EXPECT().GetByID(gomock.Any(), int64(9)).Return(domain.User{ID: 9, Username: "john"}, nil)

		useCase := usecase.NewUserUseCase(repository, storage)

		if err := useCase.DeleteAvatar(context.Background(), 9); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockuserRepository(ctrl)
		storage := gen.NewMockavatarStorage(ctrl)

		avatarURL := "http://storage/avatars/users/9/file.jpg"
		repository.EXPECT().GetByID(gomock.Any(), int64(9)).Return(domain.User{
			ID:        9,
			Username:  "john",
			AvatarURL: &avatarURL,
		}, nil)
		storage.EXPECT().DeleteAvatar(gomock.Any(), avatarURL).Return(nil)
		repository.EXPECT().ClearAvatarURL(gomock.Any(), int64(9)).Return(nil)

		useCase := usecase.NewUserUseCase(repository, storage)

		if err := useCase.DeleteAvatar(context.Background(), 9); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func stringPtr(value string) *string {
	return &value
}
