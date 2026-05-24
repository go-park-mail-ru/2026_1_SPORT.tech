package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func (repository *Repository) CreateNotification(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
	now := time.Now().UTC()

	const query = `
		INSERT INTO content_notification (
			user_id,
			type,
			actor_user_id,
			title,
			body,
			post_id,
			comment_id,
			donation_id,
			subscription_id,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
		ON CONFLICT DO NOTHING
		RETURNING notification_id, created_at, read_at
	`

	created := notification
	err := repository.db.QueryRowContext(
		ctx,
		query,
		notification.UserID,
		string(notification.Type),
		notification.ActorUserID,
		notification.Title,
		notification.Body,
		nullInt64(notification.PostID),
		nullInt64(notification.CommentID),
		nullInt64(notification.DonationID),
		nullInt64(notification.SubscriptionID),
		now,
	).Scan(&created.NotificationID, &created.CreatedAt, &created.ReadAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			return notification, nil
		}
		return domain.Notification{}, err
	}

	return created, nil
}

func (repository *Repository) ListNotifications(ctx context.Context, userID int64, limit int32, offset int32) ([]domain.Notification, error) {
	const query = `
		SELECT
			notification_id,
			user_id,
			type,
			actor_user_id,
			title,
			body,
			read_at,
			created_at,
			post_id,
			comment_id,
			donation_id,
			subscription_id
		FROM content_notification
		WHERE user_id = $1
		ORDER BY created_at DESC, notification_id DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := repository.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := make([]domain.Notification, 0)
	for rows.Next() {
		notification, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}

	return notifications, rows.Err()
}

func (repository *Repository) MarkNotificationRead(ctx context.Context, userID int64, notificationID int64) (domain.Notification, error) {
	const query = `
		UPDATE content_notification
		SET read_at = COALESCE(read_at, $3),
			updated_at = $3
		WHERE user_id = $1
			AND notification_id = $2
		RETURNING
			notification_id,
			user_id,
			type,
			actor_user_id,
			title,
			body,
			read_at,
			created_at,
			post_id,
			comment_id,
			donation_id,
			subscription_id
	`

	notification, err := scanNotification(repository.db.QueryRowContext(ctx, query, userID, notificationID, time.Now().UTC()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Notification{}, domain.ErrNotificationNotFound
		}
		return domain.Notification{}, err
	}

	return notification, nil
}

func scanNotification(scanner sqlScanner) (domain.Notification, error) {
	var notification domain.Notification
	var notificationType string
	var readAt sql.NullTime
	var postID sql.NullInt64
	var commentID sql.NullInt64
	var donationID sql.NullInt64
	var subscriptionID sql.NullInt64

	if err := scanner.Scan(
		&notification.NotificationID,
		&notification.UserID,
		&notificationType,
		&notification.ActorUserID,
		&notification.Title,
		&notification.Body,
		&readAt,
		&notification.CreatedAt,
		&postID,
		&commentID,
		&donationID,
		&subscriptionID,
	); err != nil {
		return domain.Notification{}, err
	}

	notification.Type = domain.NotificationType(notificationType)
	notification.ReadAt = timePtrFromNull(readAt)
	notification.PostID = int64PtrFromNull(postID)
	notification.CommentID = int64PtrFromNull(commentID)
	notification.DonationID = int64PtrFromNull(donationID)
	notification.SubscriptionID = int64PtrFromNull(subscriptionID)

	return notification, nil
}
