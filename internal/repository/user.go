package repository

import (
	"context"
	"database/sql"
	"time"
)

type UserProfile struct {
	Username  string
	FirstName string
	LastName  string
	Bio       *string
	AvatarURL *string
}

type User struct {
	ID        int64
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
	IsTrainer bool
	IsAdmin   bool
	Profile   UserProfile
}

type UserRepository struct {
	db *sql.DB
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
