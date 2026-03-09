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

func TestNewSessionService(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		authConfig config.AuthConfig
		wantErr    bool
	}{
		{
			name: "valid ttl",
			authConfig: config.AuthConfig{
				SessionTTL: "2h",
			},
			wantErr: false,
		},
		{
			name: "invalid ttl",
			authConfig: config.AuthConfig{
				SessionTTL: "not-a-duration",
			},
			wantErr: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewSessionService(&sessionRepositoryStub{}, testCase.authConfig)
			if testCase.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !testCase.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCreateSession(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("CreateSession returned error: %v", err)
	}
	if sessionID == "" {
		t.Fatal("expected non-empty session id")
	}
	if capturedParams.UserID != userID {
		t.Fatalf("unexpected user id: got %d, want %d", capturedParams.UserID, userID)
	}
	if capturedParams.SessionIDHash != hashSessionID(sessionID) {
		t.Fatal("expected repository to receive hashed session id")
	}

	minExpiresAt := before.Add(sessionTTL)
	maxExpiresAt := after.Add(sessionTTL)
	if capturedParams.ExpiresAt.Before(minExpiresAt) || capturedParams.ExpiresAt.After(maxExpiresAt) {
		t.Fatalf("unexpected expires at: got %v, want between %v and %v", capturedParams.ExpiresAt, minExpiresAt, maxExpiresAt)
	}
}

func TestCreateSessionRepositoryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("create session")
	service := &SessionService{
		sessionRepository: &sessionRepositoryStub{
			createSessionFunc: func(ctx context.Context, params repository.CreateSessionParams) error {
				return expectedErr
			},
		},
		sessionTTL: time.Hour,
	}

	_, err := service.CreateSession(context.Background(), 1)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected error: got %v, want %v", err, expectedErr)
	}
}

func TestGetUserIDBySessionID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sessionID  string
		repository *sessionRepositoryStub
		wantUserID int64
		wantErr    error
	}{
		{
			name:      "active session",
			sessionID: "raw-session-id",
			repository: &sessionRepositoryStub{
				getActiveSessionByHashFunc: func(ctx context.Context, sessionIDHash string) (repository.Session, error) {
					if sessionIDHash != hashSessionID("raw-session-id") {
						t.Fatalf("unexpected session hash: got %s", sessionIDHash)
					}

					return repository.Session{UserID: 7}, nil
				},
			},
			wantUserID: 7,
		},
		{
			name:      "session not found",
			sessionID: "missing-session",
			repository: &sessionRepositoryStub{
				getActiveSessionByHashFunc: func(ctx context.Context, sessionIDHash string) (repository.Session, error) {
					return repository.Session{}, sql.ErrNoRows
				},
			},
			wantErr: ErrSessionNotFound,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := &SessionService{
				sessionRepository: testCase.repository,
			}

			userID, err := service.GetUserIDBySessionID(context.Background(), testCase.sessionID)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("unexpected error: got %v, want %v", err, testCase.wantErr)
			}
			if userID != testCase.wantUserID {
				t.Fatalf("unexpected user id: got %d, want %d", userID, testCase.wantUserID)
			}
		})
	}
}

func TestRevokeSession(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("revoke session")
	testCases := []struct {
		name       string
		sessionID  string
		repository *sessionRepositoryStub
		wantErr    error
	}{
		{
			name:      "success",
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
		{
			name:      "repository error",
			sessionID: "raw-session-id",
			repository: &sessionRepositoryStub{
				revokeSessionByHashFunc: func(ctx context.Context, sessionIDHash string) error {
					return expectedErr
				},
			},
			wantErr: expectedErr,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			service := &SessionService{
				sessionRepository: testCase.repository,
			}

			err := service.RevokeSession(context.Background(), testCase.sessionID)
			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("unexpected error: got %v, want %v", err, testCase.wantErr)
			}
		})
	}
}
