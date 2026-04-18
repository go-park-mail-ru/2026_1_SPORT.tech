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
}

type SessionRepository interface {
	Create(ctx context.Context, session domain.Session) error
	GetByHash(ctx context.Context, sessionHash string) (domain.Session, error)
	RevokeByHash(ctx context.Context, sessionHash string) error
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

type AuthResult struct {
	Account          domain.Account
	SessionToken     string
	SessionExpiresAt time.Time
}

type SessionResult struct {
	Account domain.Account
	Session domain.Session
}
