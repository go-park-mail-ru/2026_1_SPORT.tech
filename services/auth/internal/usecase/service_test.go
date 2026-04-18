package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
)

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time {
	return clock.now
}

type stubPasswordHasher struct {
	hashResult string
	hashErr    error
	compareErr error
}

func (hasher stubPasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	return hasher.hashResult, hasher.hashErr
}

func (hasher stubPasswordHasher) Compare(ctx context.Context, passwordHash string, password string) error {
	return hasher.compareErr
}

type stubTokenGenerator struct {
	token string
	err   error
}

func (generator stubTokenGenerator) NewToken(ctx context.Context) (string, error) {
	return generator.token, generator.err
}

type stubAccountRepository struct {
	createFunc     func(ctx context.Context, params CreateAccountParams) (domain.Account, error)
	getByEmailFunc func(ctx context.Context, email string) (domain.Account, error)
	getByIDFunc    func(ctx context.Context, userID int64) (domain.Account, error)
}

func (repository stubAccountRepository) Create(ctx context.Context, params CreateAccountParams) (domain.Account, error) {
	return repository.createFunc(ctx, params)
}

func (repository stubAccountRepository) GetByEmail(ctx context.Context, email string) (domain.Account, error) {
	return repository.getByEmailFunc(ctx, email)
}

func (repository stubAccountRepository) GetByID(ctx context.Context, userID int64) (domain.Account, error) {
	return repository.getByIDFunc(ctx, userID)
}

type stubSessionRepository struct {
	createFunc       func(ctx context.Context, session domain.Session) error
	getByHashFunc    func(ctx context.Context, sessionHash string) (domain.Session, error)
	revokeByHashFunc func(ctx context.Context, sessionHash string) error
}

func (repository stubSessionRepository) Create(ctx context.Context, session domain.Session) error {
	return repository.createFunc(ctx, session)
}

func (repository stubSessionRepository) GetByHash(ctx context.Context, sessionHash string) (domain.Session, error) {
	return repository.getByHashFunc(ctx, sessionHash)
}

func (repository stubSessionRepository) RevokeByHash(ctx context.Context, sessionHash string) error {
	return repository.revokeByHashFunc(ctx, sessionHash)
}

func TestServiceRegister(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	accountRepositoryCalled := false
	sessionRepositoryCalled := false

	service := NewService(
		stubAccountRepository{
			createFunc: func(ctx context.Context, params CreateAccountParams) (domain.Account, error) {
				accountRepositoryCalled = true
				if params.Email != "john@example.com" {
					t.Fatalf("unexpected email: %s", params.Email)
				}
				if params.Username != "john_doe" {
					t.Fatalf("unexpected username: %s", params.Username)
				}
				if params.PasswordHash != "hashed-password" {
					t.Fatalf("unexpected password hash: %s", params.PasswordHash)
				}
				if params.Role != domain.RoleClient {
					t.Fatalf("unexpected role: %s", params.Role)
				}

				return domain.Account{
					ID:        7,
					Email:     params.Email,
					Username:  params.Username,
					Role:      params.Role,
					Status:    domain.StatusActive,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
		},
		stubSessionRepository{
			createFunc: func(ctx context.Context, session domain.Session) error {
				sessionRepositoryCalled = true
				if session.IDHash != hashSessionToken("plain-session-token") {
					t.Fatalf("unexpected session hash: %s", session.IDHash)
				}
				if session.ExpiresAt != now.Add(24*time.Hour) {
					t.Fatalf("unexpected expires at: %s", session.ExpiresAt)
				}

				return nil
			},
		},
		stubPasswordHasher{hashResult: "hashed-password"},
		stubTokenGenerator{token: "plain-session-token"},
		fixedClock{now: now},
		24*time.Hour,
	)

	result, err := service.Register(context.Background(), RegisterCommand{
		Email:    "JOHN@example.com",
		Username: "john_doe",
		Password: "supersecret123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !accountRepositoryCalled {
		t.Fatal("expected account repository to be called")
	}
	if !sessionRepositoryCalled {
		t.Fatal("expected session repository to be called")
	}
	if result.Account.ID != 7 {
		t.Fatalf("unexpected account id: %d", result.Account.ID)
	}
	if result.SessionToken != "plain-session-token" {
		t.Fatalf("unexpected session token: %s", result.SessionToken)
	}
}

func TestServiceLoginInvalidCredentials(t *testing.T) {
	service := NewService(
		stubAccountRepository{
			getByEmailFunc: func(ctx context.Context, email string) (domain.Account, error) {
				return domain.Account{ID: 7, Email: email, PasswordHash: "stored-hash", Status: domain.StatusActive}, nil
			},
		},
		stubSessionRepository{},
		stubPasswordHasher{compareErr: errors.New("password mismatch")},
		stubTokenGenerator{},
		fixedClock{now: time.Now().UTC()},
		24*time.Hour,
	)

	_, err := service.Login(context.Background(), LoginCommand{
		Email:    "john@example.com",
		Password: "wrong-pass",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("unexpected error: got %v, want %v", err, domain.ErrInvalidCredentials)
	}
}
