package postgres

import (
	"context"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"time"
)

func (repository *Repository) CreateDonation(ctx context.Context, donation domain.Donation) (domain.Donation, error) {
	now := time.Now().UTC()

	const query = `
		INSERT INTO content_donation (
			sender_user_id,
			recipient_user_id,
			amount_value,
			currency,
			message,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING donation_id, created_at
	`

	created := donation
	err := repository.db.QueryRowContext(
		ctx,
		query,
		donation.SenderUserID,
		donation.RecipientUserID,
		donation.AmountValue,
		donation.Currency,
		nullString(donation.Message),
		now,
	).Scan(&created.DonationID, &created.CreatedAt)
	if err != nil {
		return domain.Donation{}, err
	}

	return created, nil
}

func (repository *Repository) GetBalance(ctx context.Context, trainerUserID int64, currency string) (domain.Balance, error) {
	const query = `
		SELECT COALESCE(SUM(amount_value), 0)
		FROM content_donation
		WHERE recipient_user_id = $1
			AND currency = $2
	`

	balance := domain.Balance{
		TrainerUserID: trainerUserID,
		Currency:      currency,
	}
	if err := repository.db.QueryRowContext(ctx, query, trainerUserID, currency).Scan(&balance.AmountValue); err != nil {
		return domain.Balance{}, err
	}

	return balance, nil
}
