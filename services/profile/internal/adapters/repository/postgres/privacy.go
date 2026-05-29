package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

type PrivacySettingsRepository struct {
	db *sql.DB
}

func NewPrivacySettingsRepository(db *sql.DB) *PrivacySettingsRepository {
	return &PrivacySettingsRepository{db: db}
}

func (repository *PrivacySettingsRepository) Get(ctx context.Context, userID int64) (domain.PrivacySettings, error) {
	const query = `
		SELECT show_profile_in_search, allow_measurement_sharing, show_activity_status
		FROM profile_privacy_settings
		WHERE user_id = $1
	`

	var settings domain.PrivacySettings
	err := repository.db.QueryRowContext(ctx, query, userID).Scan(
		&settings.ShowProfileInSearch,
		&settings.AllowMeasurementSharing,
		&settings.ShowActivityStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DefaultPrivacySettings(), nil
		}

		return domain.PrivacySettings{}, err
	}

	return settings, nil
}

func (repository *PrivacySettingsRepository) Upsert(ctx context.Context, userID int64, settings domain.PrivacySettings) error {
	const query = `
		INSERT INTO profile_privacy_settings (
			user_id,
			show_profile_in_search,
			allow_measurement_sharing,
			show_activity_status,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			show_profile_in_search = EXCLUDED.show_profile_in_search,
			allow_measurement_sharing = EXCLUDED.allow_measurement_sharing,
			show_activity_status = EXCLUDED.show_activity_status,
			updated_at = EXCLUDED.updated_at
	`

	_, err := repository.db.ExecContext(
		ctx,
		query,
		userID,
		settings.ShowProfileInSearch,
		settings.AllowMeasurementSharing,
		settings.ShowActivityStatus,
		time.Now().UTC(),
	)
	return err
}
