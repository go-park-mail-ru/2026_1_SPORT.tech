package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var ErrEmailExists = errors.New("email exists")
var ErrUsernameExists = errors.New("username exists")
var ErrSportTypeNotFound = errors.New("sport type not found")

type UserProfile struct {
	Username  string
	FirstName string
	LastName  string
	Bio       *string
	AvatarURL *string
}

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsTrainer    bool
	IsAdmin      bool
	Profile      UserProfile
}

type UserRepository struct {
	db *sql.DB
}

type CreateClientParams struct {
	Username     string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
}

type CreateTrainerSportParams struct {
	SportTypeID     int64
	ExperienceYears int
	SportsRank      *string
}

type CreateTrainerParams struct {
	Username          string
	Email             string
	PasswordHash      string
	FirstName         string
	LastName          string
	EducationDegree   *string
	CareerSinceDate   time.Time
	TrainerSportItems []CreateTrainerSportParams
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repository *UserRepository) GetByID(ctx context.Context, userID int64) (User, error) {
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
		user      User
		bio       sql.NullString
		avatarURL sql.NullString
	)

	err := repository.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsTrainer,
		&user.IsAdmin,
		&user.Profile.FirstName,
		&user.Profile.LastName,
		&bio,
		&avatarURL,
	)
	if err != nil {
		return User{}, err
	}

	user.Profile.Username = user.Username
	if bio.Valid {
		user.Profile.Bio = &bio.String
	}
	if avatarURL.Valid {
		user.Profile.AvatarURL = &avatarURL.String
	}

	return user, nil
}

func (repository *UserRepository) CreateClient(ctx context.Context, params CreateClientParams) (int64, error) {
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
	if err := tx.QueryRowContext(
		ctx,
		createUserQuery,
		params.Email,
		params.PasswordHash,
	).Scan(&userID); err != nil {
		return 0, mapUserConflictError(err)
	}

	const createUserProfileQuery = `
		INSERT INTO user_profile (user_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := tx.ExecContext(
		ctx,
		createUserProfileQuery,
		userID,
		params.Username,
		params.FirstName,
		params.LastName,
	); err != nil {
		return 0, mapUserConflictError(err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return userID, nil
}

func (repository *UserRepository) CreateTrainer(ctx context.Context, params CreateTrainerParams) (int64, error) {
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
	if err := tx.QueryRowContext(
		ctx,
		createUserQuery,
		params.Email,
		params.PasswordHash,
	).Scan(&userID); err != nil {
		return 0, mapUserConflictError(err)
	}

	const createUserProfileQuery = `
		INSERT INTO user_profile (user_id, username, first_name, last_name)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := tx.ExecContext(
		ctx,
		createUserProfileQuery,
		userID,
		params.Username,
		params.FirstName,
		params.LastName,
	); err != nil {
		return 0, mapUserConflictError(err)
	}

	const createTrainerDetailsQuery = `
		INSERT INTO trainer_details (trainer_user_id, education_degree, career_since_date)
		VALUES ($1, $2, $3)
	`

	if _, err := tx.ExecContext(
		ctx,
		createTrainerDetailsQuery,
		userID,
		params.EducationDegree,
		params.CareerSinceDate,
	); err != nil {
		return 0, err
	}

	const createTrainerSportQuery = `
		INSERT INTO trainer_to_sport_type (trainer_id, sport_type_id, experience_years, sports_rank)
		VALUES ($1, $2, $3, $4)
	`

	for _, sport := range params.TrainerSportItems {
		if _, err := tx.ExecContext(
			ctx,
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

func (repository *UserRepository) GetByEmail(ctx context.Context, email string) (User, error) {
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
		user      User
		bio       sql.NullString
		avatarURL sql.NullString
	)

	err := repository.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsTrainer,
		&user.IsAdmin,
		&user.Profile.FirstName,
		&user.Profile.LastName,
		&bio,
		&avatarURL,
	)
	if err != nil {
		return User{}, err
	}

	user.Profile.Username = user.Username
	if bio.Valid {
		user.Profile.Bio = &bio.String
	}
	if avatarURL.Valid {
		user.Profile.AvatarURL = &avatarURL.String
	}

	return user, nil
}

func mapUserConflictError(err error) error {
	var pqError *pq.Error
	if !errors.As(err, &pqError) || pqError.Code != "23505" {
		return err
	}

	switch pqError.Constraint {
	case "user_email_key":
		return ErrEmailExists
	case "user_profile_username_key":
		return ErrUsernameExists
	case "trainer_to_sport_type_sport_type_id_fkey":
		return ErrSportTypeNotFound
	default:
		return err
	}
}
