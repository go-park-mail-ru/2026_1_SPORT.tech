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
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrSportTypeNotFound = errors.New("sport type not found")

type userRepository interface {
	GetByID(ctx context.Context, userID int64) (repository.User, error)
	GetByEmail(ctx context.Context, email string) (repository.User, error)
	CreateClient(ctx context.Context, params repository.CreateClientParams) (int64, error)
	CreateTrainer(ctx context.Context, params repository.CreateTrainerParams) (int64, error)
}

type RegisterClientParams struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterTrainerSportParams struct {
	SportTypeID     int64
	ExperienceYears int
	SportsRank      *string
}

type RegisterTrainerParams struct {
	Username        string
	Email           string
	Password        string
	FirstName       string
	LastName        string
	EducationDegree *string
	CareerSinceDate time.Time
	Sports          []RegisterTrainerSportParams
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

	return mapUser(user), nil
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

func (service *UserService) RegisterTrainer(ctx context.Context, params RegisterTrainerParams) (User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	sports := make([]repository.CreateTrainerSportParams, 0, len(params.Sports))
	for _, sport := range params.Sports {
		sports = append(sports, repository.CreateTrainerSportParams{
			SportTypeID:     sport.SportTypeID,
			ExperienceYears: sport.ExperienceYears,
			SportsRank:      sport.SportsRank,
		})
	}

	userID, err := service.userRepository.CreateTrainer(ctx, repository.CreateTrainerParams{
		Username:          params.Username,
		Email:             params.Email,
		PasswordHash:      string(passwordHash),
		FirstName:         params.FirstName,
		LastName:          params.LastName,
		EducationDegree:   params.EducationDegree,
		CareerSinceDate:   params.CareerSinceDate,
		TrainerSportItems: sports,
	})
	if err != nil {
		if errors.Is(err, repository.ErrEmailExists) {
			return User{}, ErrEmailExists
		}
		if errors.Is(err, repository.ErrUsernameExists) {
			return User{}, ErrUsernameExists
		}
		if errors.Is(err, repository.ErrSportTypeNotFound) {
			return User{}, ErrSportTypeNotFound
		}

		return User{}, err
	}

	return service.GetByID(ctx, userID)
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

	return mapUser(user), nil
}

func mapUser(user repository.User) User {
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
	}
}
