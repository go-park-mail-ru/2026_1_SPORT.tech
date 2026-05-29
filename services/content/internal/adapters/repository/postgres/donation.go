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
		) VALUES (
			$1::bigint,
			$2::bigint,
			$3::integer,
			$4::text,
			$5::text,
			$6::timestamptz
		)
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
			tier_id,
			created_at,
			updated_at
		) VALUES (
			$1::text,
			$2::text,
			$3::text,
			$4::bigint,
			$5::bigint,
			$6::integer,
			$7::text,
			$8::text,
			$9::text,
			$10::text,
			$11::bigint,
			$12::timestamptz,
			$12::timestamptz
		)
		RETURNING payment_id, created_at, updated_at
	`

	created := payment
	err := repository.db.QueryRowContext(
		ctx,
		query,
		payment.Provider,
		nullableString(payment.ProviderPaymentID),
		string(payment.Status),
		payment.SenderUserID,
		payment.RecipientUserID,
		payment.AmountValue,
		payment.Currency,
		nullString(payment.Message),
		payment.ConfirmationToken,
		nullableString(payment.ConfirmationURL),
		nullableInt64(payment.TierID),
		now,
	).Scan(&created.PaymentID, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	return created, nil
}

func (repository *Repository) ListStalePendingPayments(ctx context.Context, olderThan time.Time, limit int32) ([]domain.DonationPayment, error) {
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
			tier_id,
			donation_id,
			subscription_id,
			created_at,
			updated_at,
			confirmed_at
		FROM content_payment
		WHERE status = 'pending'
			AND provider_payment_id IS NOT NULL
			AND created_at < $1::timestamptz
		ORDER BY created_at ASC, payment_id ASC
		LIMIT $2::integer
	`

	rows, err := repository.db.QueryContext(ctx, query, olderThan, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := make([]domain.DonationPayment, 0)
	for rows.Next() {
		payment, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, rows.Err()
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
			tier_id,
			donation_id,
			subscription_id,
			created_at,
			updated_at,
			confirmed_at
		FROM content_payment
		WHERE payment_id = $1::bigint
			AND sender_user_id = $2::bigint
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

func (repository *Repository) ConfirmDonationPayment(ctx context.Context, senderUserID int64, paymentID int64, confirmationToken string) (domain.DonationPayment, bool, error) {
	return repository.confirmPayment(ctx, func(ctx context.Context, tx *sql.Tx) (domain.DonationPayment, error) {
		payment, err := scanPayment(tx.QueryRowContext(ctx, selectPaymentForUpdateByIDQuery, paymentID))
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
		return payment, nil
	}, "")
}

func (repository *Repository) ConfirmPaymentByProviderID(ctx context.Context, providerPaymentID string, stripeSubscriptionID string) (domain.DonationPayment, bool, error) {
	return repository.confirmPayment(ctx, func(ctx context.Context, tx *sql.Tx) (domain.DonationPayment, error) {
		payment, err := scanPayment(tx.QueryRowContext(ctx, selectPaymentForUpdateByProviderIDQuery, providerPaymentID))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.DonationPayment{}, domain.ErrPaymentNotFound
			}
			return domain.DonationPayment{}, err
		}
		return payment, nil
	}, stripeSubscriptionID)
}

func (repository *Repository) SetPaymentStatusByProviderID(ctx context.Context, providerPaymentID string, status domain.PaymentStatus) (domain.DonationPayment, bool, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DonationPayment{}, false, err
	}
	defer tx.Rollback()

	payment, err := scanPayment(tx.QueryRowContext(ctx, selectPaymentForUpdateByProviderIDQuery, providerPaymentID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DonationPayment{}, false, domain.ErrPaymentNotFound
		}
		return domain.DonationPayment{}, false, err
	}

	if payment.Status.IsTerminal() {
		if err := tx.Commit(); err != nil {
			return domain.DonationPayment{}, false, err
		}
		return payment, false, nil
	}

	now := time.Now().UTC()
	if _, err := tx.ExecContext(
		ctx,
		`
			UPDATE content_payment
			SET status = $2::text,
				updated_at = $3::timestamptz
			WHERE payment_id = $1::bigint
		`,
		payment.PaymentID,
		string(status),
		now,
	); err != nil {
		return domain.DonationPayment{}, false, err
	}

	payment.Status = status
	payment.UpdatedAt = now

	if err := tx.Commit(); err != nil {
		return domain.DonationPayment{}, false, err
	}

	return payment, true, nil
}

func (repository *Repository) confirmPayment(
	ctx context.Context,
	lockPayment func(ctx context.Context, tx *sql.Tx) (domain.DonationPayment, error),
	stripeSubscriptionID string,
) (domain.DonationPayment, bool, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DonationPayment{}, false, err
	}
	defer tx.Rollback()

	payment, err := lockPayment(ctx, tx)
	if err != nil {
		return domain.DonationPayment{}, false, err
	}

	if payment.Status == domain.PaymentStatusConfirmed {
		if payment.Donation != nil {
			donation, err := repository.getDonationTx(ctx, tx, payment.Donation.DonationID)
			if err != nil {
				return domain.DonationPayment{}, false, err
			}
			payment.Donation = &donation
		}
		if payment.Subscription != nil {
			subscription, err := repository.getSubscriptionTx(ctx, tx, payment.Subscription.SubscriptionID)
			if err != nil {
				return domain.DonationPayment{}, false, err
			}
			payment.Subscription = &subscription
		}
		if err := tx.Commit(); err != nil {
			return domain.DonationPayment{}, false, err
		}
		return payment, false, nil
	}

	now := time.Now().UTC()
	if payment.TierID == nil {
		createdDonation, err := createDonationTx(ctx, tx, domain.Donation{
			SenderUserID:    payment.SenderUserID,
			RecipientUserID: payment.RecipientUserID,
			AmountValue:     payment.AmountValue,
			Currency:        payment.Currency,
			Message:         payment.Message,
		}, now)
		if err != nil {
			return domain.DonationPayment{}, false, err
		}
		if _, err := tx.ExecContext(
			ctx,
			`
				UPDATE content_payment
				SET status = $2::text,
					donation_id = $3::bigint,
					confirmed_at = $4::timestamptz,
					updated_at = $4::timestamptz
				WHERE payment_id = $1::bigint
			`,
			payment.PaymentID,
			string(domain.PaymentStatusConfirmed),
			createdDonation.DonationID,
			now,
		); err != nil {
			return domain.DonationPayment{}, false, err
		}
		payment.Donation = &createdDonation
	} else {
		createdSubscription, err := subscribeToTrainerTx(ctx, tx, domain.Subscription{
			ClientUserID:  payment.SenderUserID,
			TrainerUserID: payment.RecipientUserID,
			TierID:        *payment.TierID,
			ExpiresAt:     now.AddDate(0, 1, 0),
		}, now)
		if err != nil {
			return domain.DonationPayment{}, false, err
		}
		if stripeSubscriptionID != "" {
			periodEnd := createdSubscription.ExpiresAt
			if _, err := tx.ExecContext(
				ctx,
				`
					UPDATE content_subscription
					SET stripe_subscription_id = $2::text,
						current_period_end = $3::timestamptz,
						auto_renew = TRUE,
						updated_at = $4::timestamptz
					WHERE subscription_id = $1::bigint
				`,
				createdSubscription.SubscriptionID,
				stripeSubscriptionID,
				periodEnd,
				now,
			); err != nil {
				return domain.DonationPayment{}, false, err
			}
			createdSubscription.StripeSubscriptionID = stripeSubscriptionID
			createdSubscription.CurrentPeriodEnd = &periodEnd
			createdSubscription.AutoRenew = true
		}
		if _, err := tx.ExecContext(
			ctx,
			`
				UPDATE content_payment
				SET status = $2::text,
					subscription_id = $3::bigint,
					confirmed_at = $4::timestamptz,
					updated_at = $4::timestamptz
				WHERE payment_id = $1::bigint
			`,
			payment.PaymentID,
			string(domain.PaymentStatusConfirmed),
			createdSubscription.SubscriptionID,
			now,
		); err != nil {
			return domain.DonationPayment{}, false, err
		}
		payment.Subscription = &createdSubscription
	}

	payment.Status = domain.PaymentStatusConfirmed
	payment.ConfirmedAt = &now
	payment.UpdatedAt = now

	if err := tx.Commit(); err != nil {
		return domain.DonationPayment{}, false, err
	}

	return payment, true, nil
}

const selectPaymentForUpdateByIDQuery = `
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
		tier_id,
		donation_id,
		subscription_id,
		created_at,
		updated_at,
		confirmed_at
	FROM content_payment
	WHERE payment_id = $1::bigint
	FOR UPDATE
`

const selectPaymentForUpdateByProviderIDQuery = `
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
		tier_id,
		donation_id,
		subscription_id,
		created_at,
		updated_at,
		confirmed_at
	FROM content_payment
	WHERE provider_payment_id = $1::text
	FOR UPDATE
`

func (repository *Repository) GetBalance(ctx context.Context, trainerUserID int64, currency string) (domain.Balance, error) {
	const query = `
		SELECT
			COALESCE((
				SELECT SUM(amount_value)
				FROM content_donation
				WHERE recipient_user_id = $1::bigint
					AND currency = $2::text
			), 0)
			+
			COALESCE((
				SELECT SUM(amount_value)
				FROM content_payment
				WHERE recipient_user_id = $1::bigint
					AND currency = $2::text
					AND status = 'confirmed'
					AND tier_id IS NOT NULL
			), 0)
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
		) VALUES (
			$1::bigint,
			$2::bigint,
			$3::integer,
			$4::text,
			$5::text,
			$6::timestamptz
		)
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
		WHERE donation_id = $1::bigint
	`

	return scanDonation(tx.QueryRowContext(ctx, query, donationID))
}

func subscribeToTrainerTx(ctx context.Context, tx *sql.Tx, subscription domain.Subscription, now time.Time) (domain.Subscription, error) {
	var subscriptionID int64
	err := tx.QueryRowContext(
		ctx,
		`
			SELECT subscription_id
			FROM content_subscription
			WHERE client_user_id = $1::bigint
				AND trainer_user_id = $2::bigint
				AND active = TRUE
			FOR UPDATE
		`,
		subscription.ClientUserID,
		subscription.TrainerUserID,
	).Scan(&subscriptionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.Subscription{}, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return scanSubscription(tx.QueryRowContext(
			ctx,
			`
				WITH inserted AS (
					INSERT INTO content_subscription (
						client_user_id,
						trainer_user_id,
						tier_id,
						active,
						expires_at,
						created_at,
						updated_at
					)
					VALUES (
						$1::bigint,
						$2::bigint,
						$3::integer,
						TRUE,
						$4::timestamptz,
						$5::timestamptz,
						$5::timestamptz
					)
					RETURNING subscription_id, client_user_id, trainer_user_id, tier_id, active, expires_at, created_at, updated_at, stripe_subscription_id, current_period_end, auto_renew
				)
				SELECT
					inserted.subscription_id,
					inserted.client_user_id,
					inserted.trainer_user_id,
					inserted.tier_id,
					tier.name,
					tier.price,
					inserted.active,
					inserted.expires_at,
					inserted.created_at,
					inserted.updated_at, inserted.stripe_subscription_id, inserted.current_period_end, inserted.auto_renew
				FROM inserted
				JOIN content_subscription_tier tier
					ON tier.trainer_user_id = inserted.trainer_user_id
					AND tier.tier_id = inserted.tier_id
			`,
			subscription.ClientUserID,
			subscription.TrainerUserID,
			subscription.TierID,
			subscription.ExpiresAt,
			now,
		))
	}

	return scanSubscription(tx.QueryRowContext(
		ctx,
		`
			WITH updated AS (
				UPDATE content_subscription
				SET tier_id = $1::integer,
					active = TRUE,
					expires_at = $2::timestamptz,
					updated_at = $3::timestamptz
				WHERE subscription_id = $4::bigint
				RETURNING subscription_id, client_user_id, trainer_user_id, tier_id, active, expires_at, created_at, updated_at, stripe_subscription_id, current_period_end, auto_renew
			)
			SELECT
				updated.subscription_id,
				updated.client_user_id,
				updated.trainer_user_id,
				updated.tier_id,
				tier.name,
				tier.price,
				updated.active,
				updated.expires_at,
				updated.created_at,
				updated.updated_at, updated.stripe_subscription_id, updated.current_period_end, updated.auto_renew
			FROM updated
			JOIN content_subscription_tier tier
				ON tier.trainer_user_id = updated.trainer_user_id
				AND tier.tier_id = updated.tier_id
			`,
		subscription.TierID,
		subscription.ExpiresAt,
		now,
		subscriptionID,
	))
}

func (repository *Repository) getSubscriptionTx(ctx context.Context, tx *sql.Tx, subscriptionID int64) (domain.Subscription, error) {
	const query = `
		SELECT
			subscription.subscription_id,
			subscription.client_user_id,
			subscription.trainer_user_id,
			subscription.tier_id,
			tier.name,
			tier.price,
			(subscription.active AND subscription.expires_at > now()) AS active,
			subscription.expires_at,
			subscription.created_at,
			subscription.updated_at, subscription.stripe_subscription_id, subscription.current_period_end, subscription.auto_renew
		FROM content_subscription subscription
		JOIN content_subscription_tier tier
			ON tier.trainer_user_id = subscription.trainer_user_id
			AND tier.tier_id = subscription.tier_id
		WHERE subscription.subscription_id = $1::bigint
	`

	return scanSubscription(tx.QueryRowContext(ctx, query, subscriptionID))
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
	var tierID sql.NullInt64
	var donationID sql.NullInt64
	var subscriptionID sql.NullInt64
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
		&tierID,
		&donationID,
		&subscriptionID,
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
	if tierID.Valid {
		payment.TierID = &tierID.Int64
	}
	if donationID.Valid {
		payment.Donation = &domain.Donation{DonationID: donationID.Int64}
	}
	if subscriptionID.Valid {
		payment.Subscription = &domain.Subscription{SubscriptionID: subscriptionID.Int64}
	}
	payment.ConfirmedAt = timePtrFromNull(confirmedAt)

	return payment, nil
}

func (repository *Repository) GetTrainerStatistics(ctx context.Context, trainerUserID int64, currency string, monthStart time.Time) (domain.TrainerStatistics, error) {
	const query = `
		SELECT
			(SELECT COUNT(*)::int FROM content_post WHERE author_user_id = $1),
			(SELECT COUNT(*)::int FROM content_donation WHERE recipient_user_id = $1 AND currency = $2),
			(SELECT COALESCE(SUM(amount_value), 0)::int FROM content_donation WHERE recipient_user_id = $1 AND currency = $2)
			+ (SELECT COALESCE(SUM(amount_value), 0)::int FROM content_payment WHERE recipient_user_id = $1 AND currency = $2 AND status = 'confirmed' AND tier_id IS NOT NULL),
			(
				SELECT COALESCE(SUM(amount_value), 0)::int
				FROM content_donation
				WHERE recipient_user_id = $1
					AND currency = $2
					AND created_at >= $3
			)
			+ (
				SELECT COALESCE(SUM(amount_value), 0)::int
				FROM content_payment
				WHERE recipient_user_id = $1
					AND currency = $2
					AND status = 'confirmed'
					AND tier_id IS NOT NULL
					AND confirmed_at >= $3
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

func (repository *Repository) ListReceivedDonations(ctx context.Context, recipientUserID int64, limit, offset int32) ([]domain.Donation, error) {
	const query = `
		SELECT donation_id, sender_user_id, recipient_user_id, amount_value, currency, message, created_at
		FROM content_donation
		WHERE recipient_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := repository.db.QueryContext(ctx, query, recipientUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Donation
	for rows.Next() {
		var d domain.Donation
		var msg sql.NullString
		if err := rows.Scan(&d.DonationID, &d.SenderUserID, &d.RecipientUserID,
			&d.AmountValue, &d.Currency, &msg, &d.CreatedAt); err != nil {
			return nil, err
		}
		if msg.Valid {
			d.Message = &msg.String
		}
		result = append(result, d)
	}
	return result, rows.Err()
}

func (repository *Repository) CountReceivedDonations(ctx context.Context, recipientUserID int64) (int32, error) {
	var count int32
	const query = `SELECT COUNT(*)::int FROM content_donation WHERE recipient_user_id = $1`
	if err := repository.db.QueryRowContext(ctx, query, recipientUserID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
