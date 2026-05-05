package postgres

import (
	"context"
	"database/sql"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"time"
)

func (repository *ProfileRepository) listTrainerSports(ctx context.Context, userID int64) ([]domain.TrainerSport, error) {
	const query = `
		SELECT sport_type_id, experience_years, sports_rank
		FROM trainer_sport
		WHERE user_id = $1
		ORDER BY sport_type_id
	`

	rows, err := repository.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sports := make([]domain.TrainerSport, 0)
	for rows.Next() {
		var (
			sport      domain.TrainerSport
			sportsRank sql.NullString
		)
		if err := rows.Scan(&sport.SportTypeID, &sport.ExperienceYears, &sportsRank); err != nil {
			return nil, err
		}
		if sportsRank.Valid {
			sport.SportsRank = &sportsRank.String
		}
		sports = append(sports, sport)
	}

	return sports, rows.Err()
}

func saveTrainerDetails(ctx context.Context, tx *sql.Tx, profile domain.Profile, now time.Time) error {
	const upsertTrainerProfileQuery = `
		INSERT INTO trainer_profile (
			user_id,
			education_degree,
			career_since_date,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			education_degree = EXCLUDED.education_degree,
			career_since_date = EXCLUDED.career_since_date,
			updated_at = EXCLUDED.updated_at
	`

	var educationDegree any
	var careerSinceDate any
	if profile.TrainerDetails != nil {
		educationDegree = nullString(profile.TrainerDetails.EducationDegree)
		careerSinceDate = nullTime(profile.TrainerDetails.CareerSinceDate)
	}

	if _, err := tx.ExecContext(ctx, upsertTrainerProfileQuery, profile.UserID, educationDegree, careerSinceDate, now); err != nil {
		return mapProfileError(err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM trainer_sport WHERE user_id = $1`, profile.UserID); err != nil {
		return err
	}

	if profile.TrainerDetails == nil {
		return nil
	}

	const insertSportQuery = `
		INSERT INTO trainer_sport (
			user_id,
			sport_type_id,
			experience_years,
			sports_rank,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $5)
	`

	for _, sport := range profile.TrainerDetails.Sports {
		if _, err := tx.ExecContext(
			ctx,
			insertSportQuery,
			profile.UserID,
			sport.SportTypeID,
			sport.ExperienceYears,
			nullString(sport.SportsRank),
			now,
		); err != nil {
			return mapProfileError(err)
		}
	}

	return nil
}
