package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
	"github.com/lib/pq"
)

type UserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (repository *UserRepository) GetByID(ctx context.Context, userID int64) (domain.User, error) {
	const query = `
		SELECT
			u.user_id,
			up.username,
			u.email,
			u.created_at,
			u.updated_at,
			td.trainer_user_id IS NOT NULL AS is_trainer,
			ap.admin_id IS NOT NULL AS is_admin,
			up.first_name,
			up.last_name,
			up.bio,
			up.avatar_url,
			td.education_degree,
			td.career_since_date
		FROM "user" u
		JOIN user_profile up ON up.user_id = u.user_id
		LEFT JOIN trainer_details td ON td.trainer_user_id = u.user_id
		LEFT JOIN admin_profile ap ON ap.admin_id = u.user_id
		WHERE u.user_id = $1
	`

	var (
		user            domain.User
		bio             sql.NullString
		avatarURL       sql.NullString
		educationDegree sql.NullString
		careerSinceDate sql.NullTime
	)

	err := queryRowContext(ctx, repository.db, repository.logger, "user.get_by_id", query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsTrainer,
		&user.IsAdmin,
		&user.FirstName,
		&user.LastName,
		&bio,
		&avatarURL,
		&educationDegree,
		&careerSinceDate,
	)
	if err != nil {
		return domain.User{}, err
	}

	if bio.Valid {
		user.Bio = &bio.String
	}
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}
	if user.IsTrainer {
		trainerDetails := &domain.TrainerDetails{
			Sports: make([]domain.TrainerSport, 0),
		}
		if educationDegree.Valid {
			trainerDetails.EducationDegree = &educationDegree.String
		}
		if careerSinceDate.Valid {
			trainerDetails.CareerSinceDate = careerSinceDate.Time
		}

		const sportsQuery = `
			SELECT sport_type_id, experience_years, sports_rank
			FROM trainer_to_sport_type
			WHERE trainer_id = $1
			ORDER BY sport_type_id
		`

		rows, err := queryContext(ctx, repository.db, repository.logger, "user.list_trainer_sports", sportsQuery, userID)
		if err != nil {
			return domain.User{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				sport      domain.TrainerSport
				sportsRank sql.NullString
			)

			if err := rows.Scan(
				&sport.SportTypeID,
				&sport.ExperienceYears,
				&sportsRank,
			); err != nil {
				return domain.User{}, err
			}

			if sportsRank.Valid {
				sport.SportsRank = &sportsRank.String
			}

			trainerDetails.Sports = append(trainerDetails.Sports, sport)
		}

		if err := rows.Err(); err != nil {
			return domain.User{}, err
		}

		user.TrainerDetails = trainerDetails
	}

	return user, nil
}

func (repository *UserRepository) CreateClient(ctx context.Context, command usecase.CreateClientCommand) (int64, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	const createUserQuery = `
		INSERT INTO "user" (email, password_hash)
		VALUES ($1, $2)
		RETURNING user_id
	`

	var userID int64
	if err := queryRowContext(
		ctx,
		tx,
		repository.logger,
		"user.create_client.user",
		createUserQuery,
		command.Email,
		command.PasswordHash,
	).Scan(&userID); err != nil {
		return 0, mapUserConflictError(err)
	}

	const createUserProfileQuery = `
		INSERT INTO user_profile (user_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := execContext(
		ctx,
		tx,
		repository.logger,
		"user.create_client.profile",
		createUserProfileQuery,
		userID,
		command.Username,
		command.FirstName,
		command.LastName,
	); err != nil {
		return 0, mapUserConflictError(err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}

func (repository *UserRepository) CreateTrainer(ctx context.Context, command usecase.CreateTrainerCommand) (int64, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	const createUserQuery = `
		INSERT INTO "user" (email, password_hash)
		VALUES ($1, $2)
		RETURNING user_id
	`

	var userID int64
	if err := queryRowContext(
		ctx,
		tx,
		repository.logger,
		"user.create_trainer.user",
		createUserQuery,
		command.Email,
		command.PasswordHash,
	).Scan(&userID); err != nil {
		return 0, mapUserConflictError(err)
	}

	const createUserProfileQuery = `
		INSERT INTO user_profile (user_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := execContext(
		ctx,
		tx,
		repository.logger,
		"user.create_trainer.profile",
		createUserProfileQuery,
		userID,
		command.Username,
		command.FirstName,
		command.LastName,
	); err != nil {
		return 0, mapUserConflictError(err)
	}

	const createTrainerDetailsQuery = `
		INSERT INTO trainer_details (trainer_user_id, education_degree, career_since_date)
		VALUES ($1, $2, $3)
	`

	if _, err := execContext(
		ctx,
		tx,
		repository.logger,
		"user.create_trainer.details",
		createTrainerDetailsQuery,
		userID,
		command.EducationDegree,
		command.CareerSinceDate,
	); err != nil {
		return 0, err
	}

	const createTrainerSportQuery = `
		INSERT INTO trainer_to_sport_type (trainer_id, sport_type_id, experience_years, sports_rank)
		VALUES ($1, $2, $3, $4)
	`

	for _, sport := range command.Sports {
		if _, err := execContext(
			ctx,
			tx,
			repository.logger,
			"user.create_trainer.sport",
			createTrainerSportQuery,
			userID,
			sport.SportTypeID,
			sport.ExperienceYears,
			sport.SportsRank,
		); err != nil {
			return 0, mapUserConflictError(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}

func (repository *UserRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const query = `
		SELECT
			u.user_id,
			up.username,
			u.email,
			u.password_hash,
			u.created_at,
			u.updated_at,
			td.trainer_user_id IS NOT NULL AS is_trainer,
			ap.admin_id IS NOT NULL AS is_admin,
			up.first_name,
			up.last_name,
			up.bio,
			up.avatar_url
		FROM "user" u
		JOIN user_profile up ON up.user_id = u.user_id
		LEFT JOIN trainer_details td ON td.trainer_user_id = u.user_id
		LEFT JOIN admin_profile ap ON ap.admin_id = u.user_id
		WHERE u.email = $1
	`

	var (
		user      domain.User
		bio       sql.NullString
		avatarURL sql.NullString
	)

	err := queryRowContext(ctx, repository.db, repository.logger, "user.get_by_email", query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsTrainer,
		&user.IsAdmin,
		&user.FirstName,
		&user.LastName,
		&bio,
		&avatarURL,
	)
	if err != nil {
		return domain.User{}, err
	}

	if bio.Valid {
		user.Bio = &bio.String
	}
	if avatarURL.Valid {
		user.AvatarURL = &avatarURL.String
	}

	return user, nil
}

func (repository *UserRepository) UpdateProfile(ctx context.Context, userID int64, command usecase.UpdateProfileCommand) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const getProfileForUpdateQuery = `
		SELECT username, first_name, last_name, bio
		FROM user_profile
		WHERE user_id = $1
		FOR UPDATE
	`

	var (
		currentUsername  string
		currentFirstName string
		currentLastName  string
		currentBio       sql.NullString
	)

	err = queryRowContext(
		ctx,
		tx,
		repository.logger,
		"user.get_profile_for_update",
		getProfileForUpdateQuery,
		userID,
	).Scan(
		&currentUsername,
		&currentFirstName,
		&currentLastName,
		&currentBio,
	)
	if err != nil {
		return err
	}

	updatedUsername := currentUsername
	if command.HasUsername {
		updatedUsername = command.Username
	}

	updatedFirstName := currentFirstName
	if command.HasFirstName {
		updatedFirstName = command.FirstName
	}

	updatedLastName := currentLastName
	if command.HasLastName {
		updatedLastName = command.LastName
	}

	var updatedBio any
	if currentBio.Valid {
		updatedBio = currentBio.String
	}
	if command.HasBio {
		if command.Bio == nil {
			updatedBio = nil
		} else {
			updatedBio = *command.Bio
		}
	}

	const updateProfileQuery = `
		UPDATE user_profile
		SET username = $2,
		    first_name = $3,
		    last_name = $4,
		    bio = $5,
		    updated_at = now()
		WHERE user_id = $1
	`

	if _, err := execContext(
		ctx,
		tx,
		repository.logger,
		"user.update_profile",
		updateProfileQuery,
		userID,
		updatedUsername,
		updatedFirstName,
		updatedLastName,
		updatedBio,
	); err != nil {
		return mapUserConflictError(err)
	}

	if command.HasEducationDegree || command.HasCareerSinceDate || command.HasSports {
		const getTrainerDetailsForUpdateQuery = `
			SELECT education_degree, career_since_date
			FROM trainer_details
			WHERE trainer_user_id = $1
			FOR UPDATE
		`

		var (
			currentEducationDegree sql.NullString
			currentCareerSinceDate time.Time
		)

		err = queryRowContext(
			ctx,
			tx,
			repository.logger,
			"user.get_trainer_details_for_update",
			getTrainerDetailsForUpdateQuery,
			userID,
		).Scan(
			&currentEducationDegree,
			&currentCareerSinceDate,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return usecase.ErrTrainerProfileForbidden
			}
			return err
		}

		var updatedEducationDegree any
		if currentEducationDegree.Valid {
			updatedEducationDegree = currentEducationDegree.String
		}
		if command.HasEducationDegree {
			if command.EducationDegree == nil {
				updatedEducationDegree = nil
			} else {
				updatedEducationDegree = *command.EducationDegree
			}
		}

		updatedCareerSinceDate := currentCareerSinceDate
		if command.HasCareerSinceDate {
			updatedCareerSinceDate = command.CareerSinceDate
		}

		const updateTrainerDetailsQuery = `
			UPDATE trainer_details
			SET education_degree = $2,
			    career_since_date = $3,
			    updated_at = now()
			WHERE trainer_user_id = $1
		`

		if _, err := execContext(
			ctx,
			tx,
			repository.logger,
			"user.update_trainer_details",
			updateTrainerDetailsQuery,
			userID,
			updatedEducationDegree,
			updatedCareerSinceDate,
		); err != nil {
			return err
		}

		if command.HasSports {
			const deleteTrainerSportsQuery = `
				DELETE FROM trainer_to_sport_type
				WHERE trainer_id = $1
			`

			if _, err := execContext(
				ctx,
				tx,
				repository.logger,
				"user.delete_trainer_sports",
				deleteTrainerSportsQuery,
				userID,
			); err != nil {
				return err
			}

			const insertTrainerSportQuery = `
				INSERT INTO trainer_to_sport_type (trainer_id, sport_type_id, experience_years, sports_rank)
				VALUES ($1, $2, $3, $4)
			`

			for _, sport := range command.Sports {
				if _, err := execContext(
					ctx,
					tx,
					repository.logger,
					"user.insert_trainer_sport",
					insertTrainerSportQuery,
					userID,
					sport.SportTypeID,
					sport.ExperienceYears,
					sport.SportsRank,
				); err != nil {
					return mapUserConflictError(err)
				}
			}
		}
	}

	return tx.Commit()
}

func (repository *UserRepository) UpdateAvatarURL(ctx context.Context, userID int64, avatarURL string) error {
	const query = `
		UPDATE user_profile
		SET avatar_url = $2,
		    updated_at = now()
		WHERE user_id = $1
	`

	result, err := execContext(
		ctx,
		repository.db,
		repository.logger,
		"user.update_avatar_url",
		query,
		userID,
		avatarURL,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (repository *UserRepository) ClearAvatarURL(ctx context.Context, userID int64) error {
	const query = `
		UPDATE user_profile
		SET avatar_url = NULL,
		    updated_at = now()
		WHERE user_id = $1
	`

	result, err := execContext(
		ctx,
		repository.db,
		repository.logger,
		"user.clear_avatar_url",
		query,
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
		return sql.ErrNoRows
	}

	return nil
}

func mapUserConflictError(err error) error {
	var pqError *pq.Error
	if !errors.As(err, &pqError) {
		return err
	}

	switch {
	case pqError.Code == sqlStateUniqueViolation && pqError.Constraint == userEmailUniqueConstraint:
		return usecase.ErrEmailExists
	case pqError.Code == sqlStateUniqueViolation && pqError.Constraint == userProfileUsernameUniqueConstraint:
		return usecase.ErrUsernameExists
	case pqError.Code == sqlStateForeignKeyViolation && pqError.Constraint == trainerSportTypeForeignKeyConstraint:
		return usecase.ErrSportTypeNotFound
	default:
		return err
	}
}
