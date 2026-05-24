package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func (repository *Repository) HasActiveChatSubscription(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
	var exists bool
	err := repository.db.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM content_subscription cs
				JOIN content_subscription_tier cst
					ON cst.trainer_user_id = cs.trainer_user_id
					AND cst.tier_id = cs.tier_id
				WHERE cs.client_user_id = $1
					AND cs.trainer_user_id = $2
					AND cs.active = TRUE
					AND cs.expires_at > now()
					AND cst.chat_enabled = TRUE
			)
		`,
		clientUserID,
		trainerUserID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repository *Repository) IsTrainerOf(ctx context.Context, trainerUserID int64, clientUserID int64) (bool, error) {
	var exists bool
	err := repository.db.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM content_subscription cs
				JOIN content_subscription_tier cst
					ON cst.trainer_user_id = cs.trainer_user_id
					AND cst.tier_id = cs.tier_id
				WHERE cs.trainer_user_id = $1
					AND cs.client_user_id = $2
					AND cs.active = TRUE
					AND cs.expires_at > now()
					AND cst.chat_enabled = TRUE
			)
		`,
		trainerUserID,
		clientUserID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repository *Repository) SaveChatMessage(ctx context.Context, msg domain.ChatMessage) (domain.ChatMessage, error) {
	now := time.Now().UTC()
	row := repository.db.QueryRowContext(
		ctx,
		`
			INSERT INTO content_chat_message (sender_user_id, receiver_user_id, body, created_at)
			VALUES ($1, $2, $3, $4)
			RETURNING message_id, sender_user_id, receiver_user_id, body, is_read, created_at
		`,
		msg.SenderUserID,
		msg.ReceiverUserID,
		msg.Body,
		now,
	)

	return scanChatMessage(row)
}

func (repository *Repository) ListChatMessages(ctx context.Context, userID int64, otherUserID int64, limit int32, offset int32) ([]domain.ChatMessage, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
			SELECT message_id, sender_user_id, receiver_user_id, body, is_read, created_at
			FROM content_chat_message
			WHERE (sender_user_id = $1 AND receiver_user_id = $2)
				OR (sender_user_id = $2 AND receiver_user_id = $1)
			ORDER BY created_at ASC
			LIMIT $3 OFFSET $4
		`,
		userID,
		otherUserID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]domain.ChatMessage, 0)
	for rows.Next() {
		msg, err := scanChatMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (repository *Repository) ListChatConversations(ctx context.Context, userID int64) ([]domain.ChatConversation, error) {

	rows, err := repository.db.QueryContext(
		ctx,
		`
			WITH ranked AS (
				SELECT
					CASE WHEN sender_user_id = $1 THEN receiver_user_id ELSE sender_user_id END AS other_user_id,
					message_id, sender_user_id, receiver_user_id, body, is_read, created_at,
					ROW_NUMBER() OVER (
						PARTITION BY LEAST(sender_user_id, receiver_user_id), GREATEST(sender_user_id, receiver_user_id)
						ORDER BY created_at DESC
					) AS rn
				FROM content_chat_message
				WHERE sender_user_id = $1 OR receiver_user_id = $1
			),
			unread AS (
				SELECT sender_user_id AS other_user_id, COUNT(*) AS cnt
				FROM content_chat_message
				WHERE receiver_user_id = $1 AND is_read = false
				GROUP BY sender_user_id
			)
			SELECT
				r.other_user_id,
				r.message_id, r.sender_user_id, r.receiver_user_id, r.body, r.is_read, r.created_at,
				COALESCE(u.cnt, 0) AS unread_count
			FROM ranked r
			LEFT JOIN unread u ON u.other_user_id = r.other_user_id
			WHERE r.rn = 1
			ORDER BY r.created_at DESC
		`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]domain.ChatConversation, 0)
	for rows.Next() {
		var (
			conv domain.ChatConversation
			msg  domain.ChatMessage
		)
		if err := rows.Scan(
			&conv.OtherUserID,
			&msg.MessageID,
			&msg.SenderUserID,
			&msg.ReceiverUserID,
			&msg.Body,
			&msg.IsRead,
			&msg.CreatedAt,
			&conv.UnreadCount,
		); err != nil {
			return nil, err
		}
		conv.LastMessage = msg
		conversations = append(conversations, conv)
	}

	return conversations, rows.Err()
}

func (repository *Repository) MarkChatMessageRead(ctx context.Context, userID int64, messageID int64) error {
	result, err := repository.db.ExecContext(
		ctx,
		`
			UPDATE content_chat_message
			SET is_read = true
			WHERE message_id = $1 AND receiver_user_id = $2
		`,
		messageID,
		userID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrChatMessageNotFound
	}

	return nil
}

func scanChatMessage(scanner sqlScanner) (domain.ChatMessage, error) {
	var msg domain.ChatMessage
	if err := scanner.Scan(
		&msg.MessageID,
		&msg.SenderUserID,
		&msg.ReceiverUserID,
		&msg.Body,
		&msg.IsRead,
		&msg.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ChatMessage{}, domain.ErrChatMessageNotFound
		}
		return domain.ChatMessage{}, err
	}

	return msg, nil
}
