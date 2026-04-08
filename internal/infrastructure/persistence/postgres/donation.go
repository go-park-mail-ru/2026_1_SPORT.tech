package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
	"github.com/lib/pq"
)

type DonationRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewDonationRepository(db *sql.DB, logger *slog.Logger) *DonationRepository {
	return &DonationRepository{
		db:     db,
		logger: logger,
	}
}

func (repository *DonationRepository) Create(ctx context.Context, params usecase.CreateDonationParams) (domain.Donation, error) {
	const query = `
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
	`

	var (
		donation domain.Donation
		message  sql.NullString
	)

	err := queryRowContext(
		ctx,
		repository.db,
		repository.logger,
		"donation.create",
		query,
		params.SenderUserID,
		params.RecipientUserID,
		params.AmountMantissa,
		params.AmountScale,
		params.Currency,
		params.Message,
	).Scan(
		&donation.DonationID,
		&donation.SenderUserID,
		&donation.RecipientUserID,
		&donation.AmountValue,
		&donation.Currency,
		&message,
		&donation.CreatedAt,
		&donation.UpdatedAt,
	)
	if err != nil {
		return domain.Donation{}, mapDonationError(err)
	}

	if message.Valid {
		donation.Message = &message.String
	}

	return donation, nil
}

func mapDonationError(err error) error {
	var pqError *pq.Error
	if !errors.As(err, &pqError) {
		return err
	}

	switch {
	case pqError.Code == "23503" && pqError.Constraint == "donation_recipient_user_id_fkey":
		return usecase.ErrUserNotFound
	default:
		return err
	}
}
