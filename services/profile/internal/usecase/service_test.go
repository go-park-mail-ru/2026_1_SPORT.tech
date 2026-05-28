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
	getByUsernameFunc   func(ctx context.Context, username string) (domain.Profile, error)
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

func (repository stubProfileRepository) GetByUsername(ctx context.Context, username string) (domain.Profile, error) {
	return repository.getByUsernameFunc(ctx, username)
}

func (repository stubProfileRepository) Update(ctx context.Context, profile domain.Profile) error {
	return repository.updateFunc(ctx, profile)
}

func (repository stubProfileRepository) SetTrainer(ctx context.Context, userID int64, details *domain.TrainerDetails) error {
	return nil
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

func stubRepositories(repository stubProfileRepository, sportTypes stubSportTypeRepository) Repositories {
	return Repositories{
		Profiles: repository,
		Authors:  repository,
		Avatars:  repository,
		Sports:   sportTypes,
	}
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
		stubRepositories(stubProfileRepository{
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
		}, stubSportTypeRepository{listFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil }}),
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
		stubRepositories(stubProfileRepository{
			getByIDFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
				return domain.Profile{UserID: userID, Username: "client", FirstName: "A", LastName: "B", IsTrainer: false}, nil
			},
			updateFunc: func(ctx context.Context, profile domain.Profile) error { return nil },
		}, stubSportTypeRepository{listFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil }}),
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
		stubRepositories(stubProfileRepository{
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
		}, stubSportTypeRepository{listFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil }}),
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

func TestServiceGetProfile(t *testing.T) {
	service := NewService(
		stubRepositories(stubProfileRepository{
			getByIDFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
				return domain.Profile{UserID: userID, Username: "athlete", FirstName: "Ivan", LastName: "Petrov"}, nil
			},
		}, stubSportTypeRepository{}),
		nil,
	)

	profile, err := service.GetProfile(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UserID != 5 || profile.Username != "athlete" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestServiceGetProfileInvalidID(t *testing.T) {
	service := NewService(
		stubRepositories(stubProfileRepository{}, stubSportTypeRepository{}),
		nil,
	)

	_, err := service.GetProfile(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for zero user id")
	}
}

func TestServiceGetProfileByUsername(t *testing.T) {
	service := NewService(
		stubRepositories(stubProfileRepository{
			getByUsernameFunc: func(ctx context.Context, username string) (domain.Profile, error) {
				return domain.Profile{UserID: 5, Username: username, FirstName: "Ivan", LastName: "Petrov"}, nil
			},
		}, stubSportTypeRepository{}),
		nil,
	)

	profile, err := service.GetProfileByUsername(context.Background(), "athlete")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if profile.UserID != 5 || profile.Username != "athlete" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestServiceGetProfileByUsernameInvalid(t *testing.T) {
	service := NewService(
		stubRepositories(stubProfileRepository{}, stubSportTypeRepository{}),
		nil,
	)

	_, err := service.GetProfileByUsername(context.Background(), "ab")
	if !errors.Is(err, ErrInvalidUsername) {
		t.Fatalf("expected ErrInvalidUsername, got %v", err)
	}
}

func TestServiceDeleteAvatar(t *testing.T) {
	avatarURL := "http://storage/old-avatar.png"
	deleteCalled := false
	clearCalled := false

	service := NewService(
		stubRepositories(stubProfileRepository{
			getByIDFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
				return domain.Profile{UserID: userID, Username: "john", FirstName: "John", LastName: "Doe", AvatarURL: &avatarURL}, nil
			},
			clearAvatarURLFunc: func(ctx context.Context, userID int64) error {
				clearCalled = true
				return nil
			},
		}, stubSportTypeRepository{}),
		stubAvatarStorage{
			deleteFunc: func(ctx context.Context, url string) error {
				deleteCalled = true
				if url != avatarURL {
					t.Fatalf("unexpected avatar url: %s", url)
				}
				return nil
			},
		},
	)

	err := service.DeleteAvatar(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Fatal("expected storage delete to be called")
	}
	if !clearCalled {
		t.Fatal("expected repository clear to be called")
	}
}

func TestServiceDeleteAvatarNoAvatar(t *testing.T) {
	service := NewService(
		stubRepositories(stubProfileRepository{
			getByIDFunc: func(ctx context.Context, userID int64) (domain.Profile, error) {
				return domain.Profile{UserID: userID, Username: "john", FirstName: "John", LastName: "Doe", AvatarURL: nil}, nil
			},
		}, stubSportTypeRepository{}),
		nil,
	)

	err := service.DeleteAvatar(context.Background(), 7)
	if err != nil {
		t.Fatalf("expected no error when no avatar: %v", err)
	}
}

func TestServiceListSportTypes(t *testing.T) {
	service := NewService(
		stubRepositories(stubProfileRepository{}, stubSportTypeRepository{
			listFunc: func(ctx context.Context) ([]domain.SportType, error) {
				return []domain.SportType{
					{ID: 1, Name: "Running"},
					{ID: 2, Name: "Swimming"},
				}, nil
			},
		}),
		nil,
	)

	sportTypes, err := service.ListSportTypes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sportTypes) != 2 {
		t.Fatalf("unexpected sport types count: %d", len(sportTypes))
	}
	if sportTypes[0].Name != "Running" {
		t.Fatalf("unexpected sport type: %+v", sportTypes[0])
	}
}

func TestServiceSearchAuthorsAppliesFilters(t *testing.T) {
	minExperienceYears := int32(5)
	maxExperienceYears := int32(10)
	repositoryCalled := false

	service := NewService(
		stubRepositories(stubProfileRepository{
			searchAuthorsFunc: func(ctx context.Context, query SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
				repositoryCalled = true
				if query.Query != "Анна" ||
					len(query.SportTypeIDs) != 1 ||
					query.SportTypeIDs[0] != 3001 ||
					query.MinExperienceYears == nil ||
					*query.MinExperienceYears != 5 ||
					query.MaxExperienceYears == nil ||
					*query.MaxExperienceYears != 10 ||
					!query.OnlyWithRank ||
					query.Limit != 20 ||
					query.Offset != 5 {
					t.Fatalf("unexpected search query: %+v", query)
				}

				return []domain.AuthorSummary{{UserID: 1001, Username: "coach_anna"}}, nil
			},
		}, stubSportTypeRepository{listFunc: func(ctx context.Context) ([]domain.SportType, error) { return nil, nil }}),
		nil,
	)

	authors, err := service.SearchAuthors(context.Background(), SearchAuthorsQuery{
		Query:              " Анна ",
		SportTypeIDs:       []int64{3001},
		MinExperienceYears: &minExperienceYears,
		MaxExperienceYears: &maxExperienceYears,
		OnlyWithRank:       true,
		Limit:              20,
		Offset:             5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repositoryCalled {
		t.Fatal("expected repository search to be called")
	}
	if len(authors) != 1 || authors[0].UserID != 1001 {
		t.Fatalf("unexpected authors: %+v", authors)
	}
}
