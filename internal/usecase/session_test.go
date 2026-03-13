package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

type sessionRepositoryStub struct {
	createSessionFunc          func(ctx context.Context, session domain.Session) error
	getActiveSessionByHashFunc func(ctx context.Context, sessionIDHash string) (domain.Session, error)
	revokeSessionByHashFunc    func(ctx context.Context, sessionIDHash string) error
}

func (stub *sessionRepositoryStub) CreateSession(ctx context.Context, session domain.Session) error {
	if stub.createSessionFunc == nil {
		return nil
	}

	return stub.createSessionFunc(ctx, session)
}

func (stub *sessionRepositoryStub) GetActiveSessionByHash(ctx context.Context, sessionIDHash string) (domain.Session, error) {
	if stub.getActiveSessionByHashFunc == nil {
		return domain.Session{}, nil
	}

	return stub.getActiveSessionByHashFunc(ctx, sessionIDHash)
}

func (stub *sessionRepositoryStub) RevokeSessionByHash(ctx context.Context, sessionIDHash string) error {
	if stub.revokeSessionByHashFunc == nil {
		return nil
	}

	return stub.revokeSessionByHashFunc(ctx, sessionIDHash)
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

func TestSessionUseCaseCreateSessionPositive(t *testing.T) {
	const (
		userID     int64 = 42
		sessionTTL       = 2 * time.Hour
	)

	var capturedSession domain.Session
	useCase := &SessionUseCase{
		sessionRepository: &sessionRepositoryStub{
			createSessionFunc: func(ctx context.Context, session domain.Session) error {
				capturedSession = session
				return nil
			},
		},
		sessionTTL: sessionTTL,
	}

	before := time.Now()
	sessionID, err := useCase.CreateSession(context.Background(), userID)
	after := time.Now()
	if err != nil {
		t.Fatalf("unexpected error: got %v", err)
	}
	if sessionID == "" {
		t.Fatal("expected non-empty session id")
	}
	if capturedSession.UserID != userID {
		t.Fatalf("unexpected user id: got %d, expect %d", capturedSession.UserID, userID)
	}
	if capturedSession.SessionIDHash != hashSessionID(sessionID) {
		t.Fatal("expected repository to receive hashed session id")
	}

	minExpiresAt := before.Add(sessionTTL)
	maxExpiresAt := after.Add(sessionTTL)
	if capturedSession.ExpiresAt.Before(minExpiresAt) || capturedSession.ExpiresAt.After(maxExpiresAt) {
		t.Fatalf("unexpected expires at: got %v, expect between %v and %v", capturedSession.ExpiresAt, minExpiresAt, maxExpiresAt)
	}
}

func TestSessionUseCaseCreateSessionNegative(t *testing.T) {
	expectedErr := errors.New("create session")
	tests := []struct {
		name       string
		repository *sessionRepositoryStub
		expectErr  error
	}{
		{
			name: "Ошибка репозитория",
			repository: &sessionRepositoryStub{
				createSessionFunc: func(ctx context.Context, session domain.Session) error {
					return expectedErr
				},
			},
			expectErr: expectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := &SessionUseCase{
				sessionRepository: tt.repository,
				sessionTTL:        time.Hour,
			}

			_, err := useCase.CreateSession(context.Background(), 1)
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}

func TestSessionUseCaseGetUserIDBySessionIDPositive(t *testing.T) {
	tests := []getUserIDBySessionIDTest{
		{
			name:      "Активная сессия",
			sessionID: "raw-session-id",
			repository: &sessionRepositoryStub{
				getActiveSessionByHashFunc: func(ctx context.Context, sessionIDHash string) (domain.Session, error) {
					if sessionIDHash != hashSessionID("raw-session-id") {
						t.Fatalf("unexpected session hash: got %s", sessionIDHash)
					}

					return domain.Session{UserID: 7}, nil
				},
			},
			expectID: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := &SessionUseCase{
				sessionRepository: tt.repository,
			}

			userID, err := useCase.GetUserIDBySessionID(context.Background(), tt.sessionID)
			if err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
			if userID != tt.expectID {
				t.Fatalf("unexpected user id: got %d, expect %d", userID, tt.expectID)
			}
		})
	}
}

func TestSessionUseCaseGetUserIDBySessionIDNegative(t *testing.T) {
	tests := []getUserIDBySessionIDTest{
		{
			name:      "Сессия не найдена",
			sessionID: "missing-session",
			repository: &sessionRepositoryStub{
				getActiveSessionByHashFunc: func(ctx context.Context, sessionIDHash string) (domain.Session, error) {
					return domain.Session{}, sql.ErrNoRows
				},
			},
			expectErr: ErrSessionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := &SessionUseCase{
				sessionRepository: tt.repository,
			}

			_, err := useCase.GetUserIDBySessionID(context.Background(), tt.sessionID)
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}

func TestSessionUseCaseRevokeSessionPositive(t *testing.T) {
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
			useCase := &SessionUseCase{
				sessionRepository: tt.repository,
			}

			err := useCase.RevokeSession(context.Background(), tt.sessionID)
			if err != nil {
				t.Fatalf("unexpected error: got %v", err)
			}
		})
	}
}

func TestSessionUseCaseRevokeSessionNegative(t *testing.T) {
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
			useCase := &SessionUseCase{
				sessionRepository: tt.repository,
			}

			err := useCase.RevokeSession(context.Background(), tt.sessionID)
			if !errors.Is(err, tt.expectErr) {
				t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
			}
		})
	}
}
