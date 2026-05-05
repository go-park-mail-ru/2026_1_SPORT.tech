package postgres

import (
	"context"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

func (repository *ProfileRepository) UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) error {
	const query = `
		UPDATE profile
		SET avatar_url = $2,
			updated_at = NOW()
		WHERE user_id = $1
	`

	result, err := repository.db.ExecContext(ctx, query, userID, avatarURL)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrProfileNotFound
	}

	return nil
}

func (repository *ProfileRepository) ClearAvatarURL(ctx context.Context, userID int64) error {
	const query = `
		UPDATE profile
		SET avatar_url = NULL,
			updated_at = NOW()
		WHERE user_id = $1
	`

	result, err := repository.db.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrProfileNotFound
	}

	return nil
}
