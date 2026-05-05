package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"time"
)

func (repository *Repository) ListSubscriptionTiers(ctx context.Context, trainerUserID int64) ([]domain.SubscriptionTier, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
			SELECT tier_id, trainer_user_id, name, price, description, created_at, updated_at
			FROM content_subscription_tier
			WHERE trainer_user_id = $1
			ORDER BY price ASC, tier_id ASC
		`,
		trainerUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tiers := make([]domain.SubscriptionTier, 0)
	for rows.Next() {
		tier, err := scanSubscriptionTier(rows)
		if err != nil {
			return nil, err
		}
		tiers = append(tiers, tier)
	}

	return tiers, rows.Err()
}

func (repository *Repository) GetSubscriptionTier(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error) {
	row := repository.db.QueryRowContext(
		ctx,
		`
			SELECT tier_id, trainer_user_id, name, price, description, created_at, updated_at
			FROM content_subscription_tier
			WHERE trainer_user_id = $1
				AND tier_id = $2
		`,
		trainerUserID,
		tierID,
	)

	tier, err := scanSubscriptionTier(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SubscriptionTier{}, domain.ErrSubscriptionTierNotFound
		}
		return domain.SubscriptionTier{}, err
	}

	return tier, nil
}

func (repository *Repository) CreateSubscriptionTier(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error) {
	now := time.Now().UTC()

	row := repository.db.QueryRowContext(
		ctx,
		`
			WITH next_tier AS (
				SELECT COALESCE(MAX(tier_id), 0) + 1 AS tier_id
				FROM content_subscription_tier
				WHERE trainer_user_id = $1
			)
			INSERT INTO content_subscription_tier (
				trainer_user_id,
				tier_id,
				name,
				price,
				description,
				created_at,
				updated_at
			)
			SELECT $1, next_tier.tier_id, $2, $3, $4, $5, $5
			FROM next_tier
			RETURNING tier_id, trainer_user_id, name, price, description, created_at, updated_at
		`,
		tier.TrainerUserID,
		tier.Name,
		tier.Price,
		nullString(tier.Description),
		now,
	)

	created, err := scanSubscriptionTier(row)
	if err != nil {
		return domain.SubscriptionTier{}, err
	}

	return created, nil
}

func (repository *Repository) UpdateSubscriptionTier(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error) {
	now := time.Now().UTC()

	row := repository.db.QueryRowContext(
		ctx,
		`
			UPDATE content_subscription_tier
			SET name = $3,
				price = $4,
				description = $5,
				updated_at = $6
			WHERE trainer_user_id = $1
				AND tier_id = $2
			RETURNING tier_id, trainer_user_id, name, price, description, created_at, updated_at
		`,
		tier.TrainerUserID,
		tier.TierID,
		tier.Name,
		tier.Price,
		nullString(tier.Description),
		now,
	)

	updated, err := scanSubscriptionTier(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SubscriptionTier{}, domain.ErrSubscriptionTierNotFound
		}
		return domain.SubscriptionTier{}, err
	}

	return updated, nil
}

func (repository *Repository) DeleteSubscriptionTier(ctx context.Context, trainerUserID int64, tierID int64) error {
	result, err := repository.db.ExecContext(
		ctx,
		`
			DELETE FROM content_subscription_tier
			WHERE trainer_user_id = $1
				AND tier_id = $2
		`,
		trainerUserID,
		tierID,
	)
	if err != nil {
		if isForeignKeyViolation(err) {
			return domain.ErrSubscriptionTierInUse
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrSubscriptionTierNotFound
	}

	return nil
}
