package usecase

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailExists = errors.New("email exists")
var ErrUsernameExists = errors.New("username exists")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrSportTypeNotFound = errors.New("sport type not found")

type RegisterTrainerSportCommand struct {
	SportTypeID     int64
	ExperienceYears int
	SportsRank      *string
}

type CreateClientCommand struct {
	Username     string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
}

type CreateTrainerCommand struct {
	Username        string
	Email           string
	PasswordHash    string
	FirstName       string
	LastName        string
	EducationDegree *string
	CareerSinceDate time.Time
	Sports          []RegisterTrainerSportCommand
}

type RegisterClientCommand struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterTrainerCommand struct {
	Username        string
	Email           string
	Password        string
	FirstName       string
	LastName        string
	EducationDegree *string
	CareerSinceDate time.Time
	Sports          []RegisterTrainerSportCommand
}

type UserUseCase struct {
	userRepository userRepository
}

func NewUserUseCase(userRepository userRepository) *UserUseCase {
	return &UserUseCase{
		userRepository: userRepository,
	}
}

func (useCase *UserUseCase) GetByID(ctx context.Context, userID int64) (domain.User, error) {
	user, err := useCase.userRepository.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}

		return domain.User{}, err
	}

	return user, nil
}

func (useCase *UserUseCase) RegisterClient(ctx context.Context, command RegisterClientCommand) (domain.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	userID, err := useCase.userRepository.CreateClient(ctx, CreateClientCommand{
		Username:     command.Username,
		Email:        command.Email,
		PasswordHash: string(passwordHash),
		FirstName:    command.FirstName,
		LastName:     command.LastName,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailExists):
			return domain.User{}, ErrEmailExists
		case errors.Is(err, ErrUsernameExists):
			return domain.User{}, ErrUsernameExists
		default:
			return domain.User{}, err
		}
	}

	return useCase.GetByID(ctx, userID)
}

func (useCase *UserUseCase) RegisterTrainer(ctx context.Context, command RegisterTrainerCommand) (domain.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(command.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	userID, err := useCase.userRepository.CreateTrainer(ctx, CreateTrainerCommand{
		Username:        command.Username,
		Email:           command.Email,
		PasswordHash:    string(passwordHash),
		FirstName:       command.FirstName,
		LastName:        command.LastName,
		EducationDegree: command.EducationDegree,
		CareerSinceDate: command.CareerSinceDate,
		Sports:          command.Sports,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailExists):
			return domain.User{}, ErrEmailExists
		case errors.Is(err, ErrUsernameExists):
			return domain.User{}, ErrUsernameExists
		case errors.Is(err, ErrSportTypeNotFound):
			return domain.User{}, ErrSportTypeNotFound
		default:
			return domain.User{}, err
		}
	}

	return useCase.GetByID(ctx, userID)
}

func (useCase *UserUseCase) Authenticate(ctx context.Context, email string, password string) (domain.User, error) {
	user, err := useCase.userRepository.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, ErrInvalidCredentials
		}

		return domain.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return domain.User{}, ErrInvalidCredentials
	}

	return user, nil
}
