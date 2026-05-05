package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"time"
)

type ProfileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) *ProfileRepository {
	return &ProfileRepository{db: db}
}

func (repository *ProfileRepository) Create(ctx context.Context, profile domain.Profile) error {
	now := time.Now().UTC()
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const insertProfileQuery = `
		INSERT INTO profile (
			user_id,
			username,
			first_name,
			last_name,
			bio,
			avatar_url,
			is_trainer,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`

	_, err = tx.ExecContext(
		ctx,
		insertProfileQuery,
		profile.UserID,
		profile.Username,
		profile.FirstName,
		profile.LastName,
		nullString(profile.Bio),
		nullString(profile.AvatarURL),
		profile.IsTrainer,
		now,
	)
	if err != nil {
		return mapProfileError(err)
	}

	if profile.IsTrainer {
		if err := saveTrainerDetails(ctx, tx, profile, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repository *ProfileRepository) GetByID(ctx context.Context, userID int64) (domain.Profile, error) {
	const query = `
		SELECT
			p.user_id,
			p.username,
			p.first_name,
			p.last_name,
			p.bio,
			p.avatar_url,
			p.is_trainer,
			p.created_at,
			p.updated_at,
			tp.education_degree,
			tp.career_since_date
		FROM profile p
		LEFT JOIN trainer_profile tp ON tp.user_id = p.user_id
		WHERE p.user_id = $1
	`

	loader := repository.db.QueryRowContext(ctx, query, userID)

	var (
		profile         domain.Profile
		bio             sql.NullString
		avatarURL       sql.NullString
		educationDegree sql.NullString
		careerSinceDate sql.NullTime
	)
	err := loader.Scan(
		&profile.UserID,
		&profile.Username,
		&profile.FirstName,
		&profile.LastName,
		&bio,
		&avatarURL,
		&profile.IsTrainer,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&educationDegree,
		&careerSinceDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Profile{}, domain.ErrProfileNotFound
		}
		return domain.Profile{}, err
	}

	if bio.Valid {
		profile.Bio = &bio.String
	}
	if avatarURL.Valid {
		profile.AvatarURL = &avatarURL.String
	}
	if profile.IsTrainer {
		profile.TrainerDetails = &domain.TrainerDetails{
			Sports: make([]domain.TrainerSport, 0),
		}
		if educationDegree.Valid {
			profile.TrainerDetails.EducationDegree = &educationDegree.String
		}
		if careerSinceDate.Valid {
			profile.TrainerDetails.CareerSinceDate = &careerSinceDate.Time
		}

		sports, err := repository.listTrainerSports(ctx, userID)
		if err != nil {
			return domain.Profile{}, err
		}
		profile.TrainerDetails.Sports = sports
	}

	return profile, nil
}

func (repository *ProfileRepository) Update(ctx context.Context, profile domain.Profile) error {
	now := time.Now().UTC()
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const updateProfileQuery = `
		UPDATE profile
		SET username = $2,
			first_name = $3,
			last_name = $4,
			bio = $5,
			avatar_url = $6,
			updated_at = $7
		WHERE user_id = $1
	`

	result, err := tx.ExecContext(
		ctx,
		updateProfileQuery,
		profile.UserID,
		profile.Username,
		profile.FirstName,
		profile.LastName,
		nullString(profile.Bio),
		nullString(profile.AvatarURL),
		now,
	)
	if err != nil {
		return mapProfileError(err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrProfileNotFound
	}

	if profile.IsTrainer {
		if err := saveTrainerDetails(ctx, tx, profile, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}
