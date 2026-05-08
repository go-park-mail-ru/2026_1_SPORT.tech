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
					RETURNING subscription_id, client_user_id, trainer_user_id, tier_id, active, expires_at, created_at, updated_at
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
					inserted.updated_at
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
					RETURNING subscription_id, client_user_id, trainer_user_id, tier_id, active, expires_at, created_at, updated_at
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
					updated.updated_at
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
				subscription.updated_at
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
				RETURNING subscription_id, client_user_id, trainer_user_id, tier_id, active, expires_at, created_at, updated_at
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
				updated.updated_at
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
