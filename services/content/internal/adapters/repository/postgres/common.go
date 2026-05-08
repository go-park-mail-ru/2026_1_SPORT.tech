package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/lib/pq"
)

type sqlQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type sqlScanner interface {
	Scan(dest ...any) error
}

func ensurePostOwnership(ctx context.Context, queryer sqlQueryer, postID int64, authorUserID int64) error {
	var storedAuthorUserID int64
	err := queryer.QueryRowContext(ctx, `SELECT author_user_id FROM content_post WHERE post_id = $1`, postID).Scan(&storedAuthorUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrPostNotFound
		}
		return err
	}
	if storedAuthorUserID != authorUserID {
		return domain.ErrPostForbidden
	}

	return nil
}

func scanSubscriptionTier(scanner sqlScanner) (domain.SubscriptionTier, error) {
	var (
		tier        domain.SubscriptionTier
		description sql.NullString
	)

	if err := scanner.Scan(
		&tier.TierID,
		&tier.TrainerUserID,
		&tier.Name,
		&tier.Price,
		&description,
		&tier.CreatedAt,
		&tier.UpdatedAt,
	); err != nil {
		return domain.SubscriptionTier{}, err
	}
	if description.Valid {
		tier.Description = &description.String
	}

	return tier, nil
}

func scanSubscription(scanner sqlScanner) (domain.Subscription, error) {
	var subscription domain.Subscription
	if err := scanner.Scan(
		&subscription.SubscriptionID,
		&subscription.ClientUserID,
		&subscription.TrainerUserID,
		&subscription.TierID,
		&subscription.TierName,
		&subscription.Price,
		&subscription.Active,
		&subscription.ExpiresAt,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	); err != nil {
		return domain.Subscription{}, err
	}

	return subscription, nil
}

func nullString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *value,
		Valid:  true,
	}
}

func nullInt32(value *int32) sql.NullInt32 {
	if value == nil {
		return sql.NullInt32{}
	}

	return sql.NullInt32{
		Int32: *value,
		Valid: true,
	}
}

func nullInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}

	return sql.NullInt64{
		Int64: *value,
		Valid: true,
	}
}

func isForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23503"
}
