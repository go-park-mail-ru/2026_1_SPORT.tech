package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
)

type Clock interface {
	Now() time.Time
}

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, passwordHash string, password string) error
}

type TokenGenerator interface {
	NewToken(ctx context.Context) (string, error)
}

type AccountRepository interface {
	Create(ctx context.Context, params CreateAccountParams) (domain.Account, error)
	GetByEmail(ctx context.Context, email string) (domain.Account, error)
	GetByID(ctx context.Context, userID int64) (domain.Account, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string, now time.Time) error
	UpdateEmail(ctx context.Context, userID int64, email string, now time.Time) (domain.Account, error)
	UpdateRole(ctx context.Context, userID int64, role domain.Role, now time.Time) (domain.Account, error)
	Delete(ctx context.Context, userID int64) error
}

type SessionRepository interface {
	Create(ctx context.Context, session domain.Session) error
	GetByHash(ctx context.Context, sessionHash string) (domain.Session, error)
	RevokeByHash(ctx context.Context, sessionHash string) error
	RevokeByUserID(ctx context.Context, userID int64) error
}

type CreateAccountParams struct {
	Email        string
	Username     string
	PasswordHash string
	Role         domain.Role
	Status       domain.Status
	Now          time.Time
}

type RegisterCommand struct {
	Email    string
	Username string
	Password string
	Role     domain.Role
}

type LoginCommand struct {
	Email    string
	Password string
}

type LogoutCommand struct {
	SessionToken string
}

type GetSessionQuery struct {
	SessionToken string
}

type ChangePasswordCommand struct {
	UserID          int64
	CurrentPassword string
	NewPassword     string
}

type ChangeEmailCommand struct {
	UserID          int64
	CurrentPassword string
	NewEmail        string
}

type PromoteToTrainerCommand struct {
	UserID int64
}

type LogoutAllCommand struct {
	UserID int64
}

type DeleteAccountCommand struct {
	UserID          int64
	CurrentPassword string
}

type AuthResult struct {
	Account          domain.Account
	SessionToken     string
	SessionExpiresAt time.Time
}

type SessionResult struct {
	Account domain.Account
	Session domain.Session
}
