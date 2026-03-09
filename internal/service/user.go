package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
)

var ErrUserNotFound = errors.New("user not found")

type userRepository interface {
	GetByID(ctx context.Context, userID int64) (repository.User, error)
}

type UserProfile struct {
	Username  string
	FirstName string
	LastName  string
	Bio       *string
	AvatarURL *string
}

type User struct {
	ID        int64
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsTrainer bool
	IsAdmin   bool
	Profile   UserProfile
}

type UserService struct {
	userRepository userRepository
}

func NewUserService(userRepository userRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (service *UserService) GetByID(ctx context.Context, userID int64) (User, error) {
	user, err := service.userRepository.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}

		return User{}, err
	}

	return User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		IsTrainer: user.IsTrainer,
		IsAdmin:   user.IsAdmin,
		Profile: UserProfile{
			Username:  user.Profile.Username,
			FirstName: user.Profile.FirstName,
			LastName:  user.Profile.LastName,
			Bio:       user.Profile.Bio,
			AvatarURL: user.Profile.AvatarURL,
		},
	}, nil
}
