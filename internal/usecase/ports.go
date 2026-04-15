package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

type sportTypeRepository interface {
	ListSportTypes(ctx context.Context) ([]domain.SportType, error)
}

type sessionRepository interface {
	CreateSession(ctx context.Context, session domain.Session) error
	GetActiveSessionByHash(ctx context.Context, sessionIDHash string) (domain.Session, error)
	RevokeSessionByHash(ctx context.Context, sessionIDHash string) error
}

type postRepository interface {
	ListProfilePosts(ctx context.Context, profileUserID int64, currentUserID int64) ([]domain.PostListItem, error)
	GetByID(ctx context.Context, postID int64, currentUserID int64) (domain.Post, error)
	SetLike(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error)
	DeleteLike(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error)
	Create(ctx context.Context, trainerID int64, command CreatePostCommand) (int64, error)
	Update(ctx context.Context, trainerID int64, postID int64, command UpdatePostCommand) error
	Delete(ctx context.Context, trainerID int64, postID int64) error
}

type donationRepository interface {
	Create(ctx context.Context, params CreateDonationParams) (domain.Donation, error)
}

type userRepository interface {
	GetByID(ctx context.Context, userID int64) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	CreateClient(ctx context.Context, params CreateClientCommand) (int64, error)
	CreateTrainer(ctx context.Context, params CreateTrainerCommand) (int64, error)
	UpdateProfile(ctx context.Context, userID int64, command UpdateProfileCommand) error
	UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) error
	ClearAvatarURL(ctx context.Context, userID int64) error
}

type avatarStorage interface {
	UploadAvatar(ctx context.Context, userID int64, fileName string, contentType string, file io.Reader, size int64) (string, error)
	DeleteAvatar(ctx context.Context, avatarURL string) error
}
