package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/usecase"
	"github.com/lib/pq"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (repository *AccountRepository) Create(ctx context.Context, params usecase.CreateAccountParams) (domain.Account, error) {
	const query = `
		INSERT INTO auth_user (
			email,
			username,
			password_hash,
			role,
			status,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING user_id, email, username, password_hash, role, status, created_at, updated_at
	`

	account, err := scanAccount(repository.db.QueryRowContext(
		ctx,
		query,
		params.Email,
		params.Username,
		params.PasswordHash,
		string(params.Role),
		string(params.Status),
		params.Now,
	))
	if err != nil {
		return domain.Account{}, mapAccountError(err)
	}

	return account, nil
}

func (repository *AccountRepository) GetByEmail(ctx context.Context, email string) (domain.Account, error) {
	const query = `
		SELECT user_id, email, username, password_hash, role, status, created_at, updated_at
		FROM auth_user
		WHERE email = $1
	`

	return scanAccount(repository.db.QueryRowContext(ctx, query, email))
}

func (repository *AccountRepository) GetByID(ctx context.Context, userID int64) (domain.Account, error) {
	const query = `
		SELECT user_id, email, username, password_hash, role, status, created_at, updated_at
		FROM auth_user
		WHERE user_id = $1
	`

	return scanAccount(repository.db.QueryRowContext(ctx, query, userID))
}

func (repository *AccountRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string, now time.Time) error {
	const query = `
		UPDATE auth_user
		SET password_hash = $2, updated_at = $3
		WHERE user_id = $1
	`

	result, err := repository.db.ExecContext(ctx, query, userID, passwordHash, now)
	if err != nil {
		return err
	}

	return ensureAffected(result)
}

func (repository *AccountRepository) UpdateEmail(ctx context.Context, userID int64, email string, now time.Time) (domain.Account, error) {
	const query = `
		UPDATE auth_user
		SET email = $2, updated_at = $3
		WHERE user_id = $1
		RETURNING user_id, email, username, password_hash, role, status, created_at, updated_at
	`

	account, err := scanAccount(repository.db.QueryRowContext(ctx, query, userID, email, now))
	if err != nil {
		return domain.Account{}, mapAccountError(err)
	}

	return account, nil
}

func (repository *AccountRepository) UpdateRole(ctx context.Context, userID int64, role domain.Role, now time.Time) (domain.Account, error) {
	const query = `
		UPDATE auth_user
		SET role = $2, updated_at = $3
		WHERE user_id = $1
		RETURNING user_id, email, username, password_hash, role, status, created_at, updated_at
	`

	account, err := scanAccount(repository.db.QueryRowContext(ctx, query, userID, string(role), now))
	if err != nil {
		return domain.Account{}, mapAccountError(err)
	}

	return account, nil
}

func (repository *AccountRepository) Delete(ctx context.Context, userID int64) error {
	const query = `DELETE FROM auth_user WHERE user_id = $1`

	result, err := repository.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	return ensureAffected(result)
}

func ensureAffected(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrAccountNotFound
	}

	return nil
}

func scanAccount(scanner interface {
	Scan(dest ...any) error
}) (domain.Account, error) {
	var account domain.Account
	var role string
	var accountStatus string

	err := scanner.Scan(
		&account.ID,
		&account.Email,
		&account.Username,
		&account.PasswordHash,
		&role,
		&accountStatus,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Account{}, domain.ErrAccountNotFound
		}

		return domain.Account{}, err
	}

	account.Role, err = domain.ParseRole(role)
	if err != nil {
		return domain.Account{}, err
	}

	account.Status, err = domain.ParseStatus(accountStatus)
	if err != nil {
		return domain.Account{}, err
	}

	return account, nil
}

func mapAccountError(err error) error {
	var postgresError *pq.Error
	if !errors.As(err, &postgresError) {
		return err
	}

	if postgresError.Code != "23505" {
		return err
	}

	switch postgresError.Constraint {
	case "auth_user_email_key":
		return domain.ErrEmailTaken
	case "auth_user_username_key":
		return domain.ErrUsernameTaken
	default:
		return err
	}
}
