package handler

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

type sportTypeUseCase interface {
	ListSportTypes(ctx context.Context) ([]domain.SportType, error)
}

type sessionUseCase interface {
	CreateSession(ctx context.Context, userID int64) (string, error)
	GetUserIDBySessionID(ctx context.Context, sessionID string) (int64, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type userUseCase interface {
	GetByID(ctx context.Context, userID int64) (domain.User, error)
	RegisterClient(ctx context.Context, command usecase.RegisterClientCommand) (domain.User, error)
	RegisterTrainer(ctx context.Context, command usecase.RegisterTrainerCommand) (domain.User, error)
	Authenticate(ctx context.Context, email string, password string) (domain.User, error)
}

type postUseCase interface {
	ListProfilePosts(ctx context.Context, profileUserID int64, currentUserID int64) ([]domain.PostListItem, error)
	GetByID(ctx context.Context, postID int64, currentUserID int64) (domain.Post, error)
	Create(ctx context.Context, trainerID int64, command usecase.CreatePostCommand) (domain.Post, error)
}
