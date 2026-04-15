package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

func TestSessionRepositoryCreateSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewSessionRepository(db, nil)
	session := domain.Session{
		SessionIDHash: "hash",
		UserID:        7,
		ExpiresAt:     time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO session (session_id_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`)).
		WithArgs(session.SessionIDHash, session.UserID, session.ExpiresAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.CreateSession(context.Background(), session); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSessionRepositoryGetActiveSessionByHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewSessionRepository(db, nil)

	rows := sqlmock.NewRows([]string{"user_id"}).AddRow(int64(11))
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT user_id
		FROM session
		WHERE session_id_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > now()
	`)).
		WithArgs("hash").
		WillReturnRows(rows)

	session, err := repository.GetActiveSessionByHash(context.Background(), "hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.UserID != 11 {
		t.Fatalf("unexpected session: %+v", session)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSessionRepositoryRevokeSessionByHash(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewSessionRepository(db, nil)

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE session
		SET revoked_at = now(), updated_at = now()
		WHERE session_id_hash = $1
		  AND revoked_at IS NULL
	`)).
		WithArgs("hash").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.RevokeSessionByHash(context.Background(), "hash"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestLoggedRowScanLogsAndReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"user_id"}).RowError(0, sql.ErrNoRows)
	mock.ExpectQuery("SELECT 1").WillReturnRows(rows)

	row := queryRowContext(context.Background(), db, nil, "test.scan", "SELECT 1")

	var userID int64
	if err := row.Scan(&userID); err == nil {
		t.Fatal("expected error")
	}
}
