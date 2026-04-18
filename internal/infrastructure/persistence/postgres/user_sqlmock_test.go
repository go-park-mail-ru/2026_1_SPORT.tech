package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
	"github.com/lib/pq"
)

func TestUserRepositoryGetByIDWithTrainerDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db, nil)
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	careerSince := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	userRows := sqlmock.NewRows([]string{
		"user_id", "username", "email", "created_at", "updated_at", "is_trainer", "is_admin",
		"first_name", "last_name", "bio", "avatar_url", "education_degree", "career_since_date",
	}).AddRow(
		int64(7), "coach", "coach@example.com", now, now, true, false,
		"John", "Doe", "bio", "http://example.com/avatar.jpg", "Bachelor", careerSince,
	)
	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs(int64(7)).
		WillReturnRows(userRows)

	sportsRows := sqlmock.NewRows([]string{"sport_type_id", "experience_years", "sports_rank"}).
		AddRow(int64(1), 5, "КМС")
	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT sport_type_id, experience_years, sports_rank
			FROM trainer_to_sport_type
			WHERE trainer_id = $1
			ORDER BY sport_type_id
		`)).
		WithArgs(int64(7)).
		WillReturnRows(sportsRows)

	user, err := repository.GetByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.TrainerDetails == nil || len(user.TrainerDetails.Sports) != 1 {
		t.Fatalf("unexpected trainer details: %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepositoryListTrainers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db, nil)
	careerSince := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"user_id", "username", "first_name", "last_name", "bio", "avatar_url",
		"education_degree", "career_since_date", "sport_type_id", "experience_years", "sports_rank",
	}).
		AddRow(int64(7), "coach", "John", "Doe", "bio", "http://example.com/avatar.jpg", "Bachelor", careerSince, int64(1), int64(5), "КМС").
		AddRow(int64(7), "coach", "John", "Doe", "bio", "http://example.com/avatar.jpg", "Bachelor", careerSince, int64(2), int64(3), nil)

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			u.user_id,
			up.username,
			up.first_name,
			up.last_name,
			up.bio,
			up.avatar_url,
			td.education_degree,
			td.career_since_date,
			tts.sport_type_id,
			tts.experience_years,
			tts.sports_rank
		FROM "user" u
		JOIN user_profile up ON up.user_id = u.user_id
		JOIN trainer_details td ON td.trainer_user_id = u.user_id
		LEFT JOIN trainer_to_sport_type tts ON tts.trainer_id = u.user_id
		ORDER BY u.user_id DESC, tts.sport_type_id
	`)).WillReturnRows(rows)

	trainers, err := repository.ListTrainers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(trainers) != 1 || trainers[0].TrainerDetails == nil || len(trainers[0].TrainerDetails.Sports) != 2 {
		t.Fatalf("unexpected trainers: %+v", trainers)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepositoryGetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db, nil)
	now := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"user_id", "username", "email", "password_hash", "created_at", "updated_at", "is_trainer", "is_admin",
		"first_name", "last_name", "bio", "avatar_url",
	}).AddRow(int64(3), "john", "john@example.com", "hash", now, now, false, false, "John", "Doe", nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(`
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
	`)).
		WithArgs("john@example.com").
		WillReturnRows(rows)

	user, err := repository.GetByEmail(context.Background(), "john@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "john@example.com" || user.PasswordHash != "hash" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestUserRepositoryUpdateProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db, nil)

	mock.ExpectBegin()
	currentRows := sqlmock.NewRows([]string{"username", "first_name", "last_name", "bio"}).
		AddRow("oldname", "John", "Doe", "old bio")
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT username, first_name, last_name, bio
		FROM user_profile
		WHERE user_id = $1
		FOR UPDATE
	`)).
		WithArgs(int64(5)).
		WillReturnRows(currentRows)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE user_profile
		SET username = $2,
		    first_name = $3,
		    last_name = $4,
		    bio = $5,
		    updated_at = now()
		WHERE user_id = $1
	`)).
		WithArgs(int64(5), "newname", "John", "Doe", "new bio").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	bio := "new bio"
	err = repository.UpdateProfile(context.Background(), 5, usecase.UpdateProfileCommand{
		HasUsername: true,
		Username:    "newname",
		HasBio:      true,
		Bio:         &bio,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepositoryUpdateProfileTrainerDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db, nil)
	careerSinceDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	educationDegree := "Bachelor"
	sportsRank := "КМС"

	mock.ExpectBegin()
	currentRows := sqlmock.NewRows([]string{"username", "first_name", "last_name", "bio"}).
		AddRow("oldname", "John", "Doe", "old bio")
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT username, first_name, last_name, bio
		FROM user_profile
		WHERE user_id = $1
		FOR UPDATE
	`)).
		WithArgs(int64(5)).
		WillReturnRows(currentRows)
	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE user_profile
		SET username = $2,
		    first_name = $3,
		    last_name = $4,
		    bio = $5,
		    updated_at = now()
		WHERE user_id = $1
	`)).
		WithArgs(int64(5), "oldname", "John", "Doe", "old bio").
		WillReturnResult(sqlmock.NewResult(1, 1))
	trainerRows := sqlmock.NewRows([]string{"education_degree", "career_since_date"}).
		AddRow("Old degree", careerSinceDate)
	mock.ExpectQuery(regexp.QuoteMeta(`
			SELECT education_degree, career_since_date
			FROM trainer_details
			WHERE trainer_user_id = $1
			FOR UPDATE
		`)).
		WithArgs(int64(5)).
		WillReturnRows(trainerRows)
	mock.ExpectExec(regexp.QuoteMeta(`
			UPDATE trainer_details
			SET education_degree = $2,
			    career_since_date = $3,
			    updated_at = now()
			WHERE trainer_user_id = $1
		`)).
		WithArgs(int64(5), educationDegree, careerSinceDate).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`
				DELETE FROM trainer_to_sport_type
				WHERE trainer_id = $1
			`)).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`
				INSERT INTO trainer_to_sport_type (trainer_id, sport_type_id, experience_years, sports_rank)
				VALUES ($1, $2, $3, $4)
			`)).
		WithArgs(int64(5), int64(1), 3, &sportsRank).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repository.UpdateProfile(context.Background(), 5, usecase.UpdateProfileCommand{
		HasEducationDegree: true,
		EducationDegree:    &educationDegree,
		HasCareerSinceDate: true,
		CareerSinceDate:    careerSinceDate,
		HasSports:          true,
		Sports: []usecase.RegisterTrainerSportCommand{{
			SportTypeID:     1,
			ExperienceYears: 3,
			SportsRank:      &sportsRank,
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUserRepositoryUpdateAvatarURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db, nil)

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE user_profile
		SET avatar_url = $2,
		    updated_at = now()
		WHERE user_id = $1
	`)).
		WithArgs(int64(5), "http://example.com/avatar.jpg").
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.UpdateAvatarURL(context.Background(), 5, "http://example.com/avatar.jpg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserRepositoryUpdateAvatarURLReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db, nil)

	mock.ExpectExec("UPDATE user_profile").
		WithArgs(int64(5), "http://example.com/avatar.jpg").
		WillReturnResult(sqlmock.NewResult(1, 0))

	err = repository.UpdateAvatarURL(context.Background(), 5, "http://example.com/avatar.jpg")
	if err != sql.ErrNoRows {
		t.Fatalf("unexpected error: got %v, expect %v", err, sql.ErrNoRows)
	}
}

func TestUserRepositoryClearAvatarURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewUserRepository(db, nil)

	mock.ExpectExec(regexp.QuoteMeta(`
		UPDATE user_profile
		SET avatar_url = NULL,
		    updated_at = now()
		WHERE user_id = $1
	`)).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repository.ClearAvatarURL(context.Background(), 5); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMapUserConflictError(t *testing.T) {
	tests := []struct {
		err    error
		expect error
	}{
		{err: &pq.Error{Code: sqlStateUniqueViolation, Constraint: userEmailUniqueConstraint}, expect: usecase.ErrEmailExists},
		{err: &pq.Error{Code: sqlStateUniqueViolation, Constraint: userProfileUsernameUniqueConstraint}, expect: usecase.ErrUsernameExists},
		{err: &pq.Error{Code: sqlStateForeignKeyViolation, Constraint: trainerSportTypeForeignKeyConstraint}, expect: usecase.ErrSportTypeNotFound},
	}

	for _, tt := range tests {
		if got := mapUserConflictError(tt.err); got != tt.expect {
			t.Fatalf("unexpected mapped error: got %v, expect %v", got, tt.expect)
		}
	}
}
