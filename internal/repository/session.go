package repository

import (
	"context"
	"database/sql"
	"time"
)

type Session struct {
	UserID int64
}

type CreateSessionParams struct {
	SessionIDHash string
	UserID        int64
	ExpiresAt     time.Time
}

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

func (repository *SessionRepository) CreateSession(ctx context.Context, params CreateSessionParams) error {
	const query = `
		INSERT INTO session (session_id_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := repository.db.ExecContext(
		ctx,
		query,
		params.SessionIDHash,
		params.UserID,
		params.ExpiresAt,
	)

	return err
}

func (repository *SessionRepository) GetActiveSessionByHash(ctx context.Context, sessionIDHash string) (Session, error) {
	const query = `
		SELECT user_id
		FROM session
		WHERE session_id_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > now()
	`

	var session Session
	err := repository.db.QueryRowContext(ctx, query, sessionIDHash).Scan(&session.UserID)
	if err != nil {
		return Session{}, err
	}

	return session, nil
}

func (repository *SessionRepository) RevokeSessionByHash(ctx context.Context, sessionIDHash string) error {
	const query = `
		UPDATE session
		SET revoked_at = now(), updated_at = now()
		WHERE session_id_hash = $1
		  AND revoked_at IS NULL
	`

	_, err := repository.db.ExecContext(ctx, query, sessionIDHash)
	return err
}
