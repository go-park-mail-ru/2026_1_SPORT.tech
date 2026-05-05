package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	"github.com/lib/pq"
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

func (repository *ProfileRepository) SearchAuthors(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
	const baseQuery = `
		WITH filtered_authors AS (
			SELECT p.user_id
			FROM profile p
			WHERE p.is_trainer = TRUE
	`
	const selectQuery = `
		)
		SELECT
			p.user_id,
			p.username,
			p.first_name,
			p.last_name,
			p.bio,
			p.avatar_url,
			tp.education_degree,
			tp.career_since_date,
			ts.sport_type_id,
			ts.experience_years,
			ts.sports_rank
		FROM filtered_authors fa
		JOIN profile p ON p.user_id = fa.user_id
		LEFT JOIN trainer_profile tp ON tp.user_id = p.user_id
		LEFT JOIN trainer_sport ts ON ts.user_id = p.user_id
	`

	args := []any{}
	conditions := make([]string, 0, 2)

	if trimmedQuery := strings.TrimSpace(query.Query); trimmedQuery != "" {
		args = append(args, "%"+trimmedQuery+"%")
		placeholder := fmt.Sprintf("$%d", len(args))
		conditions = append(conditions, "(p.username ILIKE "+placeholder+" OR p.first_name ILIKE "+placeholder+" OR p.last_name ILIKE "+placeholder+" OR p.bio ILIKE "+placeholder+")")
	}
	if len(query.SportTypeIDs) > 0 || query.MinExperienceYears != nil || query.MaxExperienceYears != nil || query.OnlyWithRank {
		sportConditions := []string{"filter_ts.user_id = p.user_id"}
		if len(query.SportTypeIDs) > 0 {
			args = append(args, pq.Array(query.SportTypeIDs))
			placeholder := fmt.Sprintf("$%d", len(args))
			sportConditions = append(sportConditions, "filter_ts.sport_type_id = ANY("+placeholder+")")
		}
		if query.MinExperienceYears != nil {
			args = append(args, *query.MinExperienceYears)
			placeholder := fmt.Sprintf("$%d", len(args))
			sportConditions = append(sportConditions, "filter_ts.experience_years >= "+placeholder)
		}
		if query.MaxExperienceYears != nil {
			args = append(args, *query.MaxExperienceYears)
			placeholder := fmt.Sprintf("$%d", len(args))
			sportConditions = append(sportConditions, "filter_ts.experience_years <= "+placeholder)
		}
		if query.OnlyWithRank {
			sportConditions = append(sportConditions, "filter_ts.sports_rank IS NOT NULL AND btrim(filter_ts.sports_rank) <> ''")
		}

		conditions = append(conditions, "EXISTS (SELECT 1 FROM trainer_sport filter_ts WHERE "+strings.Join(sportConditions, " AND ")+")")
	}

	queryBuilder := strings.Builder{}
	queryBuilder.WriteString(baseQuery)
	for _, condition := range conditions {
		queryBuilder.WriteString(" AND ")
		queryBuilder.WriteString(condition)
	}

	args = append(args, query.Limit, query.Offset)
	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY p.user_id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args)))
	queryBuilder.WriteString(selectQuery)
	queryBuilder.WriteString(" ORDER BY p.user_id DESC, ts.sport_type_id")

	rows, err := repository.db.QueryContext(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	authors := make([]domain.AuthorSummary, 0)
	indexByID := make(map[int64]int)
	for rows.Next() {
		var (
			author          domain.AuthorSummary
			bio             sql.NullString
			avatarURL       sql.NullString
			educationDegree sql.NullString
			careerSinceDate sql.NullTime
			sportTypeID     sql.NullInt64
			experienceYears sql.NullInt64
			sportsRank      sql.NullString
		)

		if err := rows.Scan(
			&author.UserID,
			&author.Username,
			&author.FirstName,
			&author.LastName,
			&bio,
			&avatarURL,
			&educationDegree,
			&careerSinceDate,
			&sportTypeID,
			&experienceYears,
			&sportsRank,
		); err != nil {
			return nil, err
		}

		index, ok := indexByID[author.UserID]
		if !ok {
			if bio.Valid {
				author.Bio = &bio.String
			}
			if avatarURL.Valid {
				author.AvatarURL = &avatarURL.String
			}
			author.TrainerDetails = &domain.TrainerDetails{
				Sports: make([]domain.TrainerSport, 0),
			}
			if educationDegree.Valid {
				author.TrainerDetails.EducationDegree = &educationDegree.String
			}
			if careerSinceDate.Valid {
				author.TrainerDetails.CareerSinceDate = &careerSinceDate.Time
			}

			authors = append(authors, author)
			index = len(authors) - 1
			indexByID[author.UserID] = index
		}

		if sportTypeID.Valid {
			sport := domain.TrainerSport{
				SportTypeID:     sportTypeID.Int64,
				ExperienceYears: int(experienceYears.Int64),
			}
			if sportsRank.Valid {
				sport.SportsRank = &sportsRank.String
			}
			authors[index].TrainerDetails.Sports = append(authors[index].TrainerDetails.Sports, sport)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return authors, nil
}

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

func nullString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return *value
}

func mapProfileError(err error) error {
	var postgresError *pq.Error
	if !errors.As(err, &postgresError) {
		return err
	}

	switch postgresError.Code {
	case "23505":
		switch postgresError.Constraint {
		case "profile_pkey":
			return domain.ErrProfileExists
		case "profile_username_key":
			return domain.ErrUsernameTaken
		default:
			return err
		}
	case "23503":
		switch postgresError.Constraint {
		case "trainer_sport_sport_type_id_fkey":
			return domain.ErrSportTypeNotFound
		default:
			return err
		}
	default:
		return err
	}
}
