package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func (repository *Repository) GetNotificationPreferences(ctx context.Context, userID int64) (domain.NotificationPreferences, error) {
	const query = `
		SELECT comments, likes, donations, posts, subscriptions, meetings, email_digest
		FROM notification_preferences
		WHERE user_id = $1
	`

	var preferences domain.NotificationPreferences
	err := repository.db.QueryRowContext(ctx, query, userID).Scan(
		&preferences.Comments,
		&preferences.Likes,
		&preferences.Donations,
		&preferences.Posts,
		&preferences.Subscriptions,
		&preferences.Meetings,
		&preferences.EmailDigest,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DefaultNotificationPreferences(), nil
		}

		return domain.NotificationPreferences{}, err
	}

	return preferences, nil
}

func (repository *Repository) UpsertNotificationPreferences(ctx context.Context, userID int64, preferences domain.NotificationPreferences) error {
	const query = `
		INSERT INTO notification_preferences (
			user_id, comments, likes, donations, posts, subscriptions, meetings, email_digest, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		ON CONFLICT (user_id) DO UPDATE SET
			comments = EXCLUDED.comments,
			likes = EXCLUDED.likes,
			donations = EXCLUDED.donations,
			posts = EXCLUDED.posts,
			subscriptions = EXCLUDED.subscriptions,
			meetings = EXCLUDED.meetings,
			email_digest = EXCLUDED.email_digest,
			updated_at = EXCLUDED.updated_at
	`

	_, err := repository.db.ExecContext(
		ctx,
		query,
		userID,
		preferences.Comments,
		preferences.Likes,
		preferences.Donations,
		preferences.Posts,
		preferences.Subscriptions,
		preferences.Meetings,
		preferences.EmailDigest,
		time.Now().UTC(),
	)
	return err
}
