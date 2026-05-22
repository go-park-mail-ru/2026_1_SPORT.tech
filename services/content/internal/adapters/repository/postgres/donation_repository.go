package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
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

func (repository *Repository) CreateDonationPayment(ctx context.Context, payment domain.DonationPayment) (domain.DonationPayment, error) {
	now := time.Now().UTC()

	const query = `
		INSERT INTO content_payment (
			provider,
			provider_payment_id,
			status,
			sender_user_id,
			recipient_user_id,
			amount_value,
			currency,
			message,
			confirmation_token,
			confirmation_url,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
		RETURNING payment_id, created_at, updated_at
	`

	created := payment
	err := repository.db.QueryRowContext(
		ctx,
		query,
		payment.Provider,
		payment.ProviderPaymentID,
		string(payment.Status),
		payment.SenderUserID,
		payment.RecipientUserID,
		payment.AmountValue,
		payment.Currency,
		nullString(payment.Message),
		payment.ConfirmationToken,
		nullableString(payment.ConfirmationURL),
		now,
	).Scan(&created.PaymentID, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	return created, nil
}

func (repository *Repository) UpdateDonationPaymentProvider(ctx context.Context, paymentID int64, providerPaymentID string, confirmationURL string) (domain.DonationPayment, error) {
	const query = `
		UPDATE content_payment
		SET provider_payment_id = $2,
			confirmation_url = $3,
			updated_at = $4
		WHERE payment_id = $1
		RETURNING
			payment_id,
			provider,
			provider_payment_id,
			status,
			sender_user_id,
			recipient_user_id,
			amount_value,
			currency,
			message,
			confirmation_token,
			confirmation_url,
			donation_id,
			created_at,
			updated_at,
			confirmed_at
	`

	payment, err := scanPayment(repository.db.QueryRowContext(ctx, query, paymentID, providerPaymentID, confirmationURL, time.Now().UTC()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DonationPayment{}, domain.ErrPaymentNotFound
		}
		return domain.DonationPayment{}, err
	}

	return payment, nil
}

func (repository *Repository) GetDonationPayment(ctx context.Context, senderUserID int64, paymentID int64) (domain.DonationPayment, error) {
	const query = `
		SELECT
			payment_id,
			provider,
			provider_payment_id,
			status,
			sender_user_id,
			recipient_user_id,
			amount_value,
			currency,
			message,
			confirmation_token,
			confirmation_url,
			donation_id,
			created_at,
			updated_at,
			confirmed_at
		FROM content_payment
		WHERE payment_id = $1
			AND sender_user_id = $2
	`

	payment, err := scanPayment(repository.db.QueryRowContext(ctx, query, paymentID, senderUserID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DonationPayment{}, domain.ErrPaymentNotFound
		}
		return domain.DonationPayment{}, err
	}

	return payment, nil
}

func (repository *Repository) ConfirmDonationPayment(ctx context.Context, senderUserID int64, paymentID int64, confirmationToken string) (domain.DonationPayment, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DonationPayment{}, err
	}
	defer tx.Rollback()

	payment, err := scanPayment(tx.QueryRowContext(
		ctx,
		`
			SELECT
				payment_id,
				provider,
				provider_payment_id,
				status,
				sender_user_id,
				recipient_user_id,
				amount_value,
				currency,
				message,
				confirmation_token,
				confirmation_url,
				donation_id,
				created_at,
				updated_at,
				confirmed_at
			FROM content_payment
			WHERE payment_id = $1
			FOR UPDATE
		`,
		paymentID,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DonationPayment{}, domain.ErrPaymentNotFound
		}
		return domain.DonationPayment{}, err
	}
	if payment.SenderUserID != senderUserID {
		return domain.DonationPayment{}, domain.ErrPaymentForbidden
	}
	if payment.ConfirmationToken != confirmationToken {
		return domain.DonationPayment{}, domain.ErrPaymentTokenMismatch
	}

	if payment.Status == domain.PaymentStatusConfirmed {
		if payment.Donation != nil {
			if donation, err := repository.getDonationTx(ctx, tx, payment.Donation.DonationID); err == nil {
				payment.Donation = &donation
			} else {
				return domain.DonationPayment{}, err
			}
		}
		if err := tx.Commit(); err != nil {
			return domain.DonationPayment{}, err
		}
		return payment, nil
	}

	now := time.Now().UTC()
	donation, err := createDonationTx(ctx, tx, domain.Donation{
		SenderUserID:    payment.SenderUserID,
		RecipientUserID: payment.RecipientUserID,
		AmountValue:     payment.AmountValue,
		Currency:        payment.Currency,
		Message:         payment.Message,
	}, now)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	if _, err := tx.ExecContext(
		ctx,
		`
			UPDATE content_payment
			SET status = $3,
				donation_id = $4,
				confirmed_at = $5,
				updated_at = $5
			WHERE payment_id = $1
				AND sender_user_id = $2
		`,
		payment.PaymentID,
		senderUserID,
		string(domain.PaymentStatusConfirmed),
		donation.DonationID,
		now,
	); err != nil {
		return domain.DonationPayment{}, err
	}

	payment.Status = domain.PaymentStatusConfirmed
	payment.Donation = &donation
	payment.ConfirmedAt = &now
	payment.UpdatedAt = now

	if err := tx.Commit(); err != nil {
		return domain.DonationPayment{}, err
	}

	return payment, nil
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

func createDonationTx(ctx context.Context, tx *sql.Tx, donation domain.Donation, now time.Time) (domain.Donation, error) {
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
	err := tx.QueryRowContext(
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

func (repository *Repository) getDonationTx(ctx context.Context, tx *sql.Tx, donationID int64) (domain.Donation, error) {
	const query = `
		SELECT donation_id, sender_user_id, recipient_user_id, amount_value, currency, message, created_at
		FROM content_donation
		WHERE donation_id = $1
	`

	return scanDonation(tx.QueryRowContext(ctx, query, donationID))
}

func scanDonation(scanner sqlScanner) (domain.Donation, error) {
	var donation domain.Donation
	var message sql.NullString
	if err := scanner.Scan(
		&donation.DonationID,
		&donation.SenderUserID,
		&donation.RecipientUserID,
		&donation.AmountValue,
		&donation.Currency,
		&message,
		&donation.CreatedAt,
	); err != nil {
		return domain.Donation{}, err
	}
	if message.Valid {
		donation.Message = &message.String
	}

	return donation, nil
}

func scanPayment(scanner sqlScanner) (domain.DonationPayment, error) {
	var payment domain.DonationPayment
	var status string
	var message sql.NullString
	var providerPaymentID sql.NullString
	var confirmationURL sql.NullString
	var donationID sql.NullInt64
	var confirmedAt sql.NullTime
	if err := scanner.Scan(
		&payment.PaymentID,
		&payment.Provider,
		&providerPaymentID,
		&status,
		&payment.SenderUserID,
		&payment.RecipientUserID,
		&payment.AmountValue,
		&payment.Currency,
		&message,
		&payment.ConfirmationToken,
		&confirmationURL,
		&donationID,
		&payment.CreatedAt,
		&payment.UpdatedAt,
		&confirmedAt,
	); err != nil {
		return domain.DonationPayment{}, err
	}
	payment.Status = domain.PaymentStatus(status)
	if message.Valid {
		payment.Message = &message.String
	}
	if providerPaymentID.Valid {
		payment.ProviderPaymentID = providerPaymentID.String
	}
	if confirmationURL.Valid {
		payment.ConfirmationURL = confirmationURL.String
	}
	if donationID.Valid {
		payment.Donation = &domain.Donation{DonationID: donationID.Int64}
	}
	payment.ConfirmedAt = timePtrFromNull(confirmedAt)

	return payment, nil
}

func (repository *Repository) GetTrainerStatistics(ctx context.Context, trainerUserID int64, currency string, monthStart time.Time) (domain.TrainerStatistics, error) {
	const query = `
		SELECT
			(SELECT COUNT(*)::int FROM content_post WHERE author_user_id = $1),
			(SELECT COUNT(*)::int FROM content_donation WHERE recipient_user_id = $1 AND currency = $2),
			(SELECT COALESCE(SUM(amount_value), 0)::int FROM content_donation WHERE recipient_user_id = $1 AND currency = $2),
			(
				SELECT COALESCE(SUM(amount_value), 0)::int
				FROM content_donation
				WHERE recipient_user_id = $1
					AND currency = $2
					AND created_at >= $3
			)
	`

	statistics := domain.TrainerStatistics{
		TrainerUserID: trainerUserID,
		Currency:      currency,
	}
	if err := repository.db.QueryRowContext(ctx, query, trainerUserID, currency, monthStart).Scan(
		&statistics.PostsCount,
		&statistics.DonationsCount,
		&statistics.TotalRevenue,
		&statistics.MonthlyRevenue,
	); err != nil {
		return domain.TrainerStatistics{}, err
	}

	return statistics, nil
}
