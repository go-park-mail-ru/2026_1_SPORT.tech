package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
)

type sessionRepositoryStub struct {
	createSessionFunc          func(ctx context.Context, params repository.CreateSessionParams) error
	getActiveSessionByHashFunc func(ctx context.Context, sessionIDHash string) (repository.Session, error)
	revokeSessionByHashFunc    func(ctx context.Context, sessionIDHash string) error
}

func (stub *sessionRepositoryStub) CreateSession(ctx context.Context, params repository.CreateSessionParams) error {
	if stub.createSessionFunc == nil {
		return nil
	}

	return stub.createSessionFunc(ctx, params)
}

func (stub *sessionRepositoryStub) GetActiveSessionByHash(ctx context.Context, sessionIDHash string) (repository.Session, error) {
	if stub.getActiveSessionByHashFunc == nil {
		return repository.Session{}, nil
	}

	return stub.getActiveSessionByHashFunc(ctx, sessionIDHash)
}

func (stub *sessionRepositoryStub) RevokeSessionByHash(ctx context.Context, sessionIDHash string) error {
	if stub.revokeSessionByHashFunc == nil {
		return nil
	}

	return stub.revokeSessionByHashFunc(ctx, sessionIDHash)
}

type newSessionServiceTest struct {
	name       string
	authConfig config.AuthConfig
	expectErr  bool
}

type getUserIDBySessionIDTest struct {
	name       string
	sessionID  string
	repository *sessionRepositoryStub
	expectID   int64
	expectErr  error
}

type revokeSessionTest struct {
	name       string
	sessionID  string
	repository *sessionRepositoryStub
	expectErr  error
}

func TestNewSessionServicePositive(t *testing.T) {
	tests := []newSessionServiceTest{
		{
			name: "Корректный session ttl",
			authConfig: config.AuthConfig{
				SessionTTL: "2h",
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSessionService(&sessionRepositoryStub{}, tt.authConfig)
			if tt.expectErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
		})
	}
}

func TestNewSessionServiceNegative(t *testing.T) {
	tests := []newSessionServiceTest{
		{
			name: "Некорректный session ttl",
			authConfig: config.AuthConfig{
				SessionTTL: "not-a-duration",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSessionService(&sessionRepositoryStub{}, tt.authConfig)
			if tt.expectErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
		})
	}
}

func TestSessionServiceCreateSessionPositive(t *testing.T) {
	const (
		userID     int64 = 42
		sessionTTL       = 2 * time.Hour
	)

	var capturedParams repository.CreateSessionParams
	service := &SessionService{
		sessionRepository: &sessionRepositoryStub{
			createSessionFunc: func(ctx context.Context, params repository.CreateSessionParams) error {
				capturedParams = params
				return nil
			},
		},
		sessionTTL: sessionTTL,
	}

	before := time.Now()
	sessionID, err := service.CreateSession(context.Background(), userID)
	after := time.Now()
	if err != nil {
		t.Fatalf("unexpected error: got %v", err)
	}
	if sessionID == "" {
		t.Fatal("expected non-empty session id")
	}
	if capturedParams.UserID != userID {
		t.Fatalf("unexpected user id: got %d, expect %d", capturedParams.UserID, userID)
	}
	if capturedParams.SessionIDHash != hashSessionID(sessionID) {
		t.Fatal("expected repository to receive hashed session id")
	}

	minExpiresAt := before.Add(sessionTTL)
	maxExpiresAt := after.Add(sessionTTL)
	if capturedParams.ExpiresAt.Before(minExpiresAt) || capturedParams.ExpiresAt.After(maxExpiresAt) {
		t.Fatalf("unexpected expires at: got %v, expect between %v and %v", capturedParams.ExpiresAt, minExpiresAt, maxExpiresAt)
	}
}

func TestSessionServiceCreateSessionNegative(t *testing.T) {
	expectedErr := errors.New("create session")
	tests := []struct {
		name       string
		repository *sessionRepositoryStub
		expectErr  error
	}{
		{
			name: "Ошибка репозитория",
			repository: &sessionRepositoryStub{
				createSessionFunc: func(ctx context.Context, params repository.CreateSessionParams) error {
					return expectedErr
				},
			},
			expectErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SessionService{
				sessionRepository: tt.repository,
				sessionTTL:        time.Hour,
			}

			_, err := service.CreateSession(context.Background(), 1)
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}

func TestSessionServiceGetUserIDBySessionIDPositive(t *testing.T) {
	tests := []getUserIDBySessionIDTest{
		{
			name:      "Активная сессия",
			sessionID: "raw-session-id",
			repository: &sessionRepositoryStub{
				getActiveSessionByHashFunc: func(ctx context.Context, sessionIDHash string) (repository.Session, error) {
					if sessionIDHash != hashSessionID("raw-session-id") {
						t.Fatalf("unexpected session hash: got %s", sessionIDHash)
					}

					return repository.Session{UserID: 7}, nil
				},
			},
			expectID: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SessionService{
				sessionRepository: tt.repository,
			}

			userID, err := service.GetUserIDBySessionID(context.Background(), tt.sessionID)
			if err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
			if userID != tt.expectID {
				t.Fatalf("unexpected user id: got %d, expect %d", userID, tt.expectID)
			}
		})
	}
}

func TestSessionServiceGetUserIDBySessionIDNegative(t *testing.T) {
	tests := []getUserIDBySessionIDTest{
		{
			name:      "Сессия не найдена",
			sessionID: "missing-session",
			repository: &sessionRepositoryStub{
				getActiveSessionByHashFunc: func(ctx context.Context, sessionIDHash string) (repository.Session, error) {
					return repository.Session{}, sql.ErrNoRows
				},
			},
			expectErr: ErrSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SessionService{
				sessionRepository: tt.repository,
			}

			_, err := service.GetUserIDBySessionID(context.Background(), tt.sessionID)
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}

func TestSessionServiceRevokeSessionPositive(t *testing.T) {
	tests := []revokeSessionTest{
		{
			name:      "Успех",
			sessionID: "raw-session-id",
			repository: &sessionRepositoryStub{
				revokeSessionByHashFunc: func(ctx context.Context, sessionIDHash string) error {
					if sessionIDHash != hashSessionID("raw-session-id") {
						t.Fatalf("unexpected session hash: got %s", sessionIDHash)
					}

					return nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SessionService{
				sessionRepository: tt.repository,
			}

			err := service.RevokeSession(context.Background(), tt.sessionID)
			if err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
		})
	}
}

func TestSessionServiceRevokeSessionNegative(t *testing.T) {
	expectedErr := errors.New("revoke session")
	tests := []revokeSessionTest{
		{
			name:      "Ошибка репозитория",
			sessionID: "raw-session-id",
			repository: &sessionRepositoryStub{
				revokeSessionByHashFunc: func(ctx context.Context, sessionIDHash string) error {
					return expectedErr
				},
			},
			expectErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &SessionService{
				sessionRepository: tt.repository,
			}

			err := service.RevokeSession(context.Background(), tt.sessionID)
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}
