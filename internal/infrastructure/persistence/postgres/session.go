package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

type SessionRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSessionRepository(db *sql.DB, logger *slog.Logger) *SessionRepository {
	return &SessionRepository{
		db:     db,
		logger: logger,
	}
}

func (repository *SessionRepository) CreateSession(ctx context.Context, session domain.Session) error {
	const query = `
		INSERT INTO session (session_id_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := execContext(
		ctx,
		repository.db,
		repository.logger,
		"session.create",
		query,
		session.SessionIDHash,
		session.UserID,
		session.ExpiresAt,
	)

	return err
}

func (repository *SessionRepository) GetActiveSessionByHash(ctx context.Context, sessionIDHash string) (domain.Session, error) {
	const query = `
		SELECT user_id
		FROM session
		WHERE session_id_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > now()
	`

	var session domain.Session
	err := queryRowContext(
		ctx,
		repository.db,
		repository.logger,
		"session.get_active_by_hash",
		query,
		sessionIDHash,
	).Scan(&session.UserID)
	if err != nil {
		return domain.Session{}, err
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

	_, err := execContext(ctx, repository.db, repository.logger, "session.revoke_by_hash", query, sessionIDHash)
	return err
}
