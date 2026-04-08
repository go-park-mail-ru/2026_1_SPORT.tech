package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

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
			up.avatar_url
		FROM "user" u
		JOIN user_profile up ON up.user_id = u.user_id
		LEFT JOIN trainer_details td ON td.trainer_user_id = u.user_id
		LEFT JOIN admin_profile ap ON ap.admin_id = u.user_id
		WHERE u.user_id = $1
	`

	var (
		user      domain.User
		bio       sql.NullString
		avatarURL sql.NullString
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

	return tx.Commit()
}

func mapUserConflictError(err error) error {
	var pqError *pq.Error
	if !errors.As(err, &pqError) {
		return err
	}

	switch {
	case pqError.Code == "23505" && pqError.Constraint == "user_email_key":
		return usecase.ErrEmailExists
	case pqError.Code == "23505" && pqError.Constraint == "user_profile_username_key":
		return usecase.ErrUsernameExists
	case pqError.Code == "23503" && pqError.Constraint == "trainer_to_sport_type_sport_type_id_fkey":
		return usecase.ErrSportTypeNotFound
	default:
		return err
	}
}
