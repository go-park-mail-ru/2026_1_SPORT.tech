package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

type Repositories struct {
	Profiles ProfileRepository
	Authors  AuthorRepository
	Avatars  AvatarRepository
	Sports   SportTypeRepository
}

type ProfileRepository interface {
	Create(ctx context.Context, profile domain.Profile) error
	GetByID(ctx context.Context, userID int64) (domain.Profile, error)
	Update(ctx context.Context, profile domain.Profile) error
}

type AuthorRepository interface {
	SearchAuthors(ctx context.Context, query SearchAuthorsQuery) ([]domain.AuthorSummary, error)
}

type AvatarRepository interface {
	GetByID(ctx context.Context, userID int64) (domain.Profile, error)
	UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) error
	ClearAvatarURL(ctx context.Context, userID int64) error
}

type SportTypeRepository interface {
	ListSportTypes(ctx context.Context) ([]domain.SportType, error)
}

type AvatarStorage interface {
	UploadAvatar(ctx context.Context, userID int64, fileName string, contentType string, file io.Reader, size int64) (string, error)
	DeleteAvatar(ctx context.Context, avatarURL string) error
}

type CreateProfileCommand struct {
	UserID         int64
	Username       string
	FirstName      string
	LastName       string
	Bio            *string
	IsTrainer      bool
	TrainerDetails *domain.TrainerDetails
}

type UpdateProfileCommand struct {
	UserID            int64
	Username          *string
	FirstName         *string
	LastName          *string
	Bio               *string
	HasBio            bool
	TrainerDetails    *domain.TrainerDetails
	HasTrainerDetails bool
}

type SearchAuthorsQuery struct {
	Query              string
	SportTypeIDs       []int64
	MinExperienceYears *int32
	MaxExperienceYears *int32
	OnlyWithRank       bool
	Limit              int32
	Offset             int32
}

type UploadAvatarCommand struct {
	UserID      int64
	FileName    string
	ContentType string
	Content     []byte
}
