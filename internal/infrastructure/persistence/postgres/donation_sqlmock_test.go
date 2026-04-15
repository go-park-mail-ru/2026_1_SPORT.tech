package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
	"github.com/lib/pq"
)

func TestDonationRepositoryCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()

		repository := NewDonationRepository(db, nil)
		now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
		message := "thanks"

		rows := sqlmock.NewRows([]string{
			"donation_id",
			"sender_user_id",
			"recipient_user_id",
			"amount_mantissa",
			"currency",
			"message",
			"created_at",
			"updated_at",
		}).AddRow(int64(10), int64(1), int64(2), int64(5000), "RUB", message, now, now)

		mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO donation (
			sender_user_id,
			recipient_user_id,
			amount_mantissa,
			amount_scale,
			currency,
			message
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			donation_id,
			sender_user_id,
			recipient_user_id,
			amount_mantissa,
			currency,
			message,
			created_at,
			updated_at
	`)).
			WithArgs(int64(1), int64(2), int64(5000), 2, "RUB", &message).
			WillReturnRows(rows)

		donation, err := repository.Create(context.Background(), usecase.CreateDonationParams{
			SenderUserID:    1,
			RecipientUserID: 2,
			AmountMantissa:  5000,
			AmountScale:     2,
			Currency:        "RUB",
			Message:         &message,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if donation.DonationID != 10 || donation.Message == nil || *donation.Message != message {
			t.Fatalf("unexpected donation: %+v", donation)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("unmet expectations: %v", err)
		}
	})

	t.Run("maps recipient fk error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock.New: %v", err)
		}
		defer db.Close()

		repository := NewDonationRepository(db, nil)

		mock.ExpectQuery("INSERT INTO donation").
			WithArgs(int64(1), int64(999), int64(5000), 2, "RUB", nil).
			WillReturnError(&pq.Error{
				Code:       "23503",
				Constraint: "donation_recipient_user_id_fkey",
			})

		_, err = repository.Create(context.Background(), usecase.CreateDonationParams{
			SenderUserID:    1,
			RecipientUserID: 999,
			AmountMantissa:  5000,
			AmountScale:     2,
			Currency:        "RUB",
		})
		if err != usecase.ErrUserNotFound {
			t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrUserNotFound)
		}
	})
}
