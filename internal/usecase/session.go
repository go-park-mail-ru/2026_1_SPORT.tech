package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

const sessionIDBytesLength = 32

var ErrSessionNotFound = errors.New("session not found")

type SessionUseCase struct {
	sessionRepository sessionRepository
	sessionTTL        time.Duration
}

func NewSessionUseCase(sessionRepository sessionRepository, sessionTTL time.Duration) *SessionUseCase {
	return &SessionUseCase{
		sessionRepository: sessionRepository,
		sessionTTL:        sessionTTL,
	}
}

func (useCase *SessionUseCase) CreateSession(ctx context.Context, userID int64) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	if err := useCase.sessionRepository.CreateSession(ctx, domain.Session{
		SessionIDHash: hashSessionID(sessionID),
		UserID:        userID,
		ExpiresAt:     time.Now().Add(useCase.sessionTTL),
	}); err != nil {
		return "", err
	}

	return sessionID, nil
}

func (useCase *SessionUseCase) GetUserIDBySessionID(ctx context.Context, sessionID string) (int64, error) {
	session, err := useCase.sessionRepository.GetActiveSessionByHash(ctx, hashSessionID(sessionID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrSessionNotFound
		}

		return 0, err
	}

	return session.UserID, nil
}

func (useCase *SessionUseCase) RevokeSession(ctx context.Context, sessionID string) error {
	return useCase.sessionRepository.RevokeSessionByHash(ctx, hashSessionID(sessionID))
}

func generateSessionID() (string, error) {
	buffer := make([]byte, sessionIDBytesLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}

func hashSessionID(sessionID string) string {
	hash := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(hash[:])
}
