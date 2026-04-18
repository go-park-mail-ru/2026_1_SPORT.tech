package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (repository *SessionRepository) Create(ctx context.Context, session domain.Session) error {
	const query = `
		INSERT INTO auth_session (
			session_id_hash,
			user_id,
			expires_at,
			revoked_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := repository.db.ExecContext(
		ctx,
		query,
		session.IDHash,
		session.UserID,
		session.ExpiresAt,
		session.RevokedAt,
		session.CreatedAt,
		session.UpdatedAt,
	)
	return err
}

func (repository *SessionRepository) GetByHash(ctx context.Context, sessionHash string) (domain.Session, error) {
	const query = `
		SELECT session_id_hash, user_id, expires_at, revoked_at, created_at, updated_at
		FROM auth_session
		WHERE session_id_hash = $1
	`

	var session domain.Session
	var revokedAt sql.NullTime
	err := repository.db.QueryRowContext(ctx, query, sessionHash).Scan(
		&session.IDHash,
		&session.UserID,
		&session.ExpiresAt,
		&revokedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, domain.ErrSessionNotFound
		}

		return domain.Session{}, err
	}

	if revokedAt.Valid {
		session.RevokedAt = &revokedAt.Time
	}

	return session, nil
}

func (repository *SessionRepository) RevokeByHash(ctx context.Context, sessionHash string) error {
	const query = `
		UPDATE auth_session
		SET revoked_at = NOW(), updated_at = NOW()
		WHERE session_id_hash = $1 AND revoked_at IS NULL
	`

	result, err := repository.db.ExecContext(ctx, query, sessionHash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}
