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
var ErrEmailExists = errors.New("email exists")
var ErrUsernameExists = errors.New("username exists")

type userRepository interface {
	GetByID(ctx context.Context, userID int64) (repository.User, error)
	CreateClient(ctx context.Context, params repository.CreateClientParams) (int64, error)
}

type RegisterClientParams struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
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

func (service *UserService) RegisterClient(ctx context.Context, params RegisterClientParams) (User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	userID, err := service.userRepository.CreateClient(ctx, repository.CreateClientParams{
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: string(passwordHash),
		FirstName:    params.FirstName,
		LastName:     params.LastName,
	})
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			return User{}, ErrEmailExists
		}
		if errors.Is(err, repository.ErrUsernameExists) {
			return User{}, ErrUsernameExists
		}

		return User{}, err
	}

	return service.GetByID(ctx, userID)
}
