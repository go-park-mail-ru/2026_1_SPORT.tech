package usecase

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

type stubProfileRepository struct {
	createFunc          func(ctx context.Context, profile domain.Profile) error
	getByIDFunc         func(ctx context.Context, userID int64) (domain.Profile, error)
	updateFunc          func(ctx context.Context, profile domain.Profile) error
	searchAuthorsFunc   func(ctx context.Context, query SearchAuthorsQuery) ([]domain.AuthorSummary, error)
	updateAvatarURLFunc func(ctx context.Context, userID int64, avatarURL string) error
	clearAvatarURLFunc  func(ctx context.Context, userID int64) error
}

func (repository stubProfileRepository) Create(ctx context.Context, profile domain.Profile) error {
	return repository.createFunc(ctx, profile)
}

func (repository stubProfileRepository) GetByID(ctx context.Context, userID int64) (domain.Profile, error) {
	return repository.getByIDFunc(ctx, userID)
}

func (repository stubProfileRepository) Update(ctx context.Context, profile domain.Profile) error {
	return repository.updateFunc(ctx, profile)
}

func (repository stubProfileRepository) SearchAuthors(ctx context.Context, query SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
	return repository.searchAuthorsFunc(ctx, query)
}

func (repository stubProfileRepository) UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) error {
	return repository.updateAvatarURLFunc(ctx, userID, avatarURL)
}

func (repository stubProfileRepository) ClearAvatarURL(ctx context.Context, userID int64) error {
	return repository.clearAvatarURLFunc(ctx, userID)
}

type stubSportTypeRepository struct {
	listFunc func(ctx context.Context) ([]domain.SportType, error)
}

func (repository stubSportTypeRepository) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	return repository.listFunc(ctx)
}

type stubAvatarStorage struct {
	uploadFunc func(ctx context.Context, userID int64, fileName string, contentType string, file io.Reader, size int64) (string, error)
	deleteFunc func(ctx context.Context, avatarURL string) error
}

func (storage stubAvatarStorage) UploadAvatar(ctx context.Context, userID int64, fileName string, contentType string, file io.Reader, size int64) (string, error) {
	return storage.uploadFunc(ctx, userID, fileName, contentType, file, size)
}

func (storage stubAvatarStorage) DeleteAvatar(ctx context.Context, avatarURL string) error {
	return storage.deleteFunc(ctx, avatarURL)
}

func TestServiceCreateProfile(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	createCalled := false

	service := NewService(
		stubProfileRepository{
			createFunc: func(ctx context.Context, profile domain.Profile) error {
				createCalled = true
				if profile.UserID != 7 || profile.Username != "coach_john" {
					t.Fatalf("unexpected profile: %+v", profile)
				}
				return nil
			},
			getByIDFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
				return domain.Profile{
					UserID:    userID,
					Username:  "coach_john",
					FirstName: "John",
					LastName:  "Doe",
					IsTrainer: true,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
		},
		stubSportTypeRepository{listFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil }},
		nil,
	)

	profile, err := service.CreateProfile(context.Background(), CreateProfileCommand{
		UserID:    7,
		Username:  "coach_john",
		FirstName: "John",
		LastName:  "Doe",
		IsTrainer: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !createCalled {
		t.Fatal("expected create to be called")
	}
	if profile.UserID != 7 {
		t.Fatalf("unexpected user id: %d", profile.UserID)
	}
}

func TestServiceUpdateProfileRejectsTrainerDetailsForClient(t *testing.T) {
	service := NewService(
		stubProfileRepository{
			getByIDFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
				return domain.Profile{UserID: userID, Username: "client", FirstName: "A", LastName: "B", IsTrainer: false}, nil
			},
			updateFunc: func(ctx context.Context, profile domain.Profile) error { return nil },
		},
		stubSportTypeRepository{listFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil }},
		nil,
	)

	_, err := service.UpdateProfile(context.Background(), UpdateProfileCommand{
		UserID:            5,
		HasTrainerDetails: true,
		TrainerDetails:    &domain.TrainerDetails{},
	})
	if !errors.Is(err, domain.ErrTrainerProfileForbidden) {
		t.Fatalf("unexpected error: got %v, want %v", err, domain.ErrTrainerProfileForbidden)
	}
}

func TestServiceUploadAvatar(t *testing.T) {
	uploaded := false
	updated := false

	service := NewService(
		stubProfileRepository{
			getByIDFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
				oldURL := "http://storage/old.png"
				return domain.Profile{UserID: userID, Username: "john", FirstName: "John", LastName: "Doe", AvatarURL: &oldURL}, nil
			},
			updateAvatarURLFunc: func(ctx context.Context, userID int64, avatarURL string) error {
				updated = true
				if avatarURL != "http://storage/new.png" {
					t.Fatalf("unexpected avatar url: %s", avatarURL)
				}
				return nil
			},
		},
		stubSportTypeRepository{listFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil }},
		stubAvatarStorage{
			uploadFunc: func(ctx context.Context, userID int64, fileName string, contentType string, file io.Reader, size int64) (string, error) {
				uploaded = true
				if userID != 7 || fileName != "avatar.png" || contentType != "image/png" {
					t.Fatalf("unexpected upload args: userID=%d fileName=%s contentType=%s", userID, fileName, contentType)
				}
				payload, err := io.ReadAll(file)
				if err != nil {
					t.Fatalf("read payload: %v", err)
				}
				if !bytes.Equal(payload, []byte("img")) || size != 3 {
					t.Fatalf("unexpected payload")
				}

				return "http://storage/new.png", nil
			},
			deleteFunc: func(ctx context.Context, avatarURL string) error { return nil },
		},
	)

	_, err := service.UploadAvatar(context.Background(), UploadAvatarCommand{
		UserID:      7,
		FileName:    "avatar.png",
		ContentType: "image/png",
		Content:     []byte("img"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !uploaded || !updated {
		t.Fatal("expected upload and update to be called")
	}
}
