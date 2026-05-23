package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

type Repositories struct {
	Profiles           ProfileRepository
	Authors            AuthorRepository
	Avatars            AvatarRepository
	Sports             SportTypeRepository
	Measurements       MeasurementRepository
	MeasurementSharing MeasurementSharingRepository
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

type MeasurementRepository interface {
	CreateMeasurement(ctx context.Context, m domain.Measurement) (domain.Measurement, error)
	ListMeasurements(ctx context.Context, userID int64, limit, offset int32) ([]domain.Measurement, error)
	DeleteMeasurement(ctx context.Context, userID, measurementID int64) error
}

type MeasurementSharingRepository interface {
	// SetSharing заменяет список тренеров, которым clientUserID разрешил доступ.
	SetSharing(ctx context.Context, clientUserID int64, trainerUserIDs []int64) error
	// GetSharing возвращает список разрешённых trainer_user_id.
	GetSharing(ctx context.Context, clientUserID int64) ([]int64, error)
	// HasAccess проверяет наличие разрешения.
	HasAccess(ctx context.Context, clientUserID, trainerUserID int64) (bool, error)
}

type CreateMeasurementCommand struct {
	UserID     int64
	MeasuredAt string // "YYYY-MM-DD"
	WeightKg   *float64
	BodyFatPct *float64
	ChestCm    *int32
	WaistCm    *int32
	HipsCm     *int32
	Notes      *string
}

type DeleteMeasurementCommand struct {
	UserID        int64
	MeasurementID int64
}

type SetMeasurementSharingCommand struct {
	ClientUserID   int64
	TrainerUserIDs []int64
}

type GetMeasurementSharingQuery struct {
	ClientUserID int64
}

type ListMeasurementsQuery struct {
	UserID int64
	Limit  int32
	Offset int32
	// ViewerUserID — кто запрашивает данные.
	// 0 или == UserID означает «сам пользователь», ограничения не применяются.
	ViewerUserID int64
}
