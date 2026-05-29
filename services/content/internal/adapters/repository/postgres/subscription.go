package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"time"
)

func (repository *Repository) GetActiveSubscriptionLevel(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error) {
	var tierID int32
	err := repository.db.QueryRowContext(
		ctx,
		`
			SELECT tier_id
			FROM content_subscription
			WHERE client_user_id = $1
				AND trainer_user_id = $2
				AND active = TRUE
				AND expires_at > now()
			ORDER BY tier_id DESC
			LIMIT 1
		`,
		clientUserID,
		trainerUserID,
	).Scan(&tierID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &tierID, nil
}

func (repository *Repository) SubscribeToTrainer(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	now := time.Now().UTC()
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Subscription{}, err
	}
	defer tx.Rollback()

	var subscriptionID int64
	err = tx.QueryRowContext(
		ctx,
		`
			SELECT subscription_id
			FROM content_subscription
			WHERE client_user_id = $1
				AND trainer_user_id = $2
				AND active = TRUE
			FOR UPDATE
		`,
		subscription.ClientUserID,
		subscription.TrainerUserID,
	).Scan(&subscriptionID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.Subscription{}, err
	}

	var row *sql.Row
	if errors.Is(err, sql.ErrNoRows) {
		row = tx.QueryRowContext(
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
					VALUES ($1, $2, $3, TRUE, $4, $5, $5)
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
		)
	} else {
		row = tx.QueryRowContext(
			ctx,
			`
				WITH updated AS (
					UPDATE content_subscription
					SET tier_id = $3,
						active = TRUE,
						expires_at = $4,
						updated_at = $5
					WHERE subscription_id = $6
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
			subscription.ClientUserID,
			subscription.TrainerUserID,
			subscription.TierID,
			subscription.ExpiresAt,
			now,
			subscriptionID,
		)
	}

	created, err := scanSubscription(row)
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.Subscription{}, domain.ErrSubscriptionTierNotFound
		}
		return domain.Subscription{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Subscription{}, err
	}

	return created, nil
}

func (repository *Repository) ListSubscriptions(ctx context.Context, clientUserID int64) ([]domain.Subscription, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
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
			WHERE subscription.client_user_id = $1
			ORDER BY active DESC, subscription.created_at DESC, subscription.subscription_id DESC
		`,
		clientUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscriptions := make([]domain.Subscription, 0)
	for rows.Next() {
		subscription, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, rows.Err()
}

func (repository *Repository) ListTrainerSubscribers(ctx context.Context, trainerUserID int64, limit int32, offset int32) ([]domain.Subscription, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
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
			WHERE subscription.trainer_user_id = $1::bigint
				AND subscription.active = TRUE
				AND subscription.expires_at > now()
			ORDER BY subscription.created_at DESC, subscription.subscription_id DESC
			LIMIT $2::integer OFFSET $3::integer
		`,
		trainerUserID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscribers := make([]domain.Subscription, 0)
	for rows.Next() {
		subscription, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subscribers = append(subscribers, subscription)
	}

	return subscribers, rows.Err()
}

func (repository *Repository) UpdateSubscription(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	now := time.Now().UTC()
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Subscription{}, err
	}
	defer tx.Rollback()

	var trainerUserID int64
	if err := tx.QueryRowContext(
		ctx,
		`
			SELECT trainer_user_id
			FROM content_subscription
			WHERE client_user_id = $1
				AND subscription_id = $2
				AND active = TRUE
			FOR UPDATE
		`,
		subscription.ClientUserID,
		subscription.SubscriptionID,
	).Scan(&trainerUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Subscription{}, domain.ErrSubscriptionNotFound
		}
		return domain.Subscription{}, err
	}

	var tierExists bool
	if err := tx.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM content_subscription_tier
				WHERE trainer_user_id = $1
					AND tier_id = $2
			)
		`,
		trainerUserID,
		subscription.TierID,
	).Scan(&tierExists); err != nil {
		return domain.Subscription{}, err
	}
	if !tierExists {
		return domain.Subscription{}, domain.ErrSubscriptionTierNotFound
	}

	row := tx.QueryRowContext(
		ctx,
		`
			WITH updated AS (
				UPDATE content_subscription
				SET tier_id = $3,
					updated_at = $4
				WHERE client_user_id = $1
					AND subscription_id = $2
					AND active = TRUE
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
		subscription.ClientUserID,
		subscription.SubscriptionID,
		subscription.TierID,
		now,
	)

	updated, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Subscription{}, domain.ErrSubscriptionNotFound
		}
		return domain.Subscription{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Subscription{}, err
	}

	return updated, nil
}

func (repository *Repository) RenewSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string, currentPeriodEnd time.Time) (bool, error) {
	result, err := repository.db.ExecContext(
		ctx,
		`
			UPDATE content_subscription
			SET active = TRUE,
				expires_at = $2::timestamptz,
				current_period_end = $2::timestamptz,
				updated_at = $3::timestamptz
			WHERE stripe_subscription_id = $1::text
		`,
		stripeSubscriptionID,
		currentPeriodEnd,
		time.Now().UTC(),
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (repository *Repository) DeactivateSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (bool, error) {
	result, err := repository.db.ExecContext(
		ctx,
		`
			UPDATE content_subscription
			SET active = FALSE,
				auto_renew = FALSE,
				updated_at = $2::timestamptz
			WHERE stripe_subscription_id = $1::text
				AND active = TRUE
		`,
		stripeSubscriptionID,
		time.Now().UTC(),
	)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

func (repository *Repository) GetSubscription(ctx context.Context, clientUserID int64, subscriptionID int64) (domain.Subscription, error) {
	const query = `
		SELECT
			subscription.subscription_id,
			subscription.client_user_id,
			subscription.trainer_user_id,
			subscription.tier_id,
			tier.name,
			tier.price,
			subscription.active,
			subscription.expires_at,
			subscription.created_at,
			subscription.updated_at, subscription.stripe_subscription_id, subscription.current_period_end, subscription.auto_renew
		FROM content_subscription subscription
		JOIN content_subscription_tier tier
			ON tier.trainer_user_id = subscription.trainer_user_id
			AND tier.tier_id = subscription.tier_id
		WHERE subscription.subscription_id = $1::bigint
			AND subscription.client_user_id = $2::bigint
	`

	subscription, err := scanSubscription(repository.db.QueryRowContext(ctx, query, subscriptionID, clientUserID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Subscription{}, domain.ErrSubscriptionNotFound
		}
		return domain.Subscription{}, err
	}

	return subscription, nil
}

func (repository *Repository) SetSubscriptionAutoRenew(ctx context.Context, clientUserID int64, subscriptionID int64, autoRenew bool) error {
	result, err := repository.db.ExecContext(
		ctx,
		`
			UPDATE content_subscription
			SET auto_renew = $3::boolean,
				updated_at = $4::timestamptz
			WHERE client_user_id = $1::bigint
				AND subscription_id = $2::bigint
		`,
		clientUserID,
		subscriptionID,
		autoRenew,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrSubscriptionNotFound
	}

	return nil
}

func (repository *Repository) DeactivateExpiredSubscriptions(ctx context.Context, expiredBefore time.Time) (int64, error) {
	result, err := repository.db.ExecContext(
		ctx,
		`
			UPDATE content_subscription
			SET active = FALSE,
				updated_at = $2::timestamptz
			WHERE active = TRUE
				AND expires_at < $1::timestamptz
		`,
		expiredBefore,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (repository *Repository) CancelSubscription(ctx context.Context, clientUserID int64, subscriptionID int64) error {
	result, err := repository.db.ExecContext(
		ctx,
		`
			UPDATE content_subscription
			SET active = FALSE,
				updated_at = $3
			WHERE client_user_id = $1
				AND subscription_id = $2
				AND active = TRUE
		`,
		clientUserID,
		subscriptionID,
		time.Now().UTC(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrSubscriptionNotFound
	}

	return nil
}
