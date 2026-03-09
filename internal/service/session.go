package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
)

const sessionIDBytesLength = 32

var ErrSessionNotFound = errors.New("session not found")

type sessionRepository interface {
	CreateSession(ctx context.Context, params repository.CreateSessionParams) error
	GetActiveSessionByHash(ctx context.Context, sessionIDHash string) (repository.Session, error)
	RevokeSessionByHash(ctx context.Context, sessionIDHash string) error
}

type SessionService struct {
	sessionRepository sessionRepository
	sessionTTL        time.Duration
}

func NewSessionService(sessionRepository sessionRepository, authConfig config.AuthConfig) (*SessionService, error) {
	sessionTTL, err := authConfig.SessionTTLDuration()
	if err != nil {
		return nil, fmt.Errorf("parse session ttl: %w", err)
	}

	return &SessionService{
		sessionRepository: sessionRepository,
		sessionTTL:        sessionTTL,
	}, nil
}

func (service *SessionService) CreateSession(ctx context.Context, userID int64) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	if err := service.sessionRepository.CreateSession(ctx, repository.CreateSessionParams{
		SessionIDHash: hashSessionID(sessionID),
		UserID:        userID,
		ExpiresAt:     time.Now().Add(service.sessionTTL),
	}); err != nil {
		return "", err
	}

	return sessionID, nil
}

func (service *SessionService) GetUserIDBySessionID(ctx context.Context, sessionID string) (int64, error) {
	session, err := service.sessionRepository.GetActiveSessionByHash(ctx, hashSessionID(sessionID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrSessionNotFound
		}

		return 0, err
	}

	return session.UserID, nil
}

func (service *SessionService) RevokeSession(ctx context.Context, sessionID string) error {
	return service.sessionRepository.RevokeSessionByHash(ctx, hashSessionID(sessionID))
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
