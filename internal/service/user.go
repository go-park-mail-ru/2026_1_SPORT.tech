package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("invalid credentials")

type userRepository interface {
	GetByID(ctx context.Context, userID int64) (repository.User, error)
	GetByEmail(ctx context.Context, email string) (repository.User, error)
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

func (service *UserService) Authenticate(ctx context.Context, email string, password string) (User, error) {
	user, err := service.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}

		return User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return User{}, ErrInvalidCredentials
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
