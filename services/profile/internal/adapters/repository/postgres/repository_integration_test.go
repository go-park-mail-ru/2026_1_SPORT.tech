package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	_ "github.com/lib/pq"
)

const repositorySchemaSQL = `
DROP TABLE IF EXISTS trainer_sport;
DROP TABLE IF EXISTS trainer_profile;
DROP TABLE IF EXISTS sport_type;
DROP TABLE IF EXISTS profile;

CREATE TABLE profile (
	user_id bigint PRIMARY KEY,
	username text NOT NULL,
	first_name text NOT NULL,
	last_name text NOT NULL,
	bio text,
	avatar_url text,
	is_trainer boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	CONSTRAINT profile_username_key UNIQUE (username)
);

CREATE TABLE trainer_profile (
	user_id bigint PRIMARY KEY REFERENCES profile(user_id) ON DELETE CASCADE,
	education_degree text,
	career_since_date date,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL
);

CREATE TABLE sport_type (
	sport_type_id bigint PRIMARY KEY,
	name text NOT NULL UNIQUE
);

CREATE TABLE trainer_sport (
	user_id bigint NOT NULL REFERENCES trainer_profile(user_id) ON DELETE CASCADE,
	sport_type_id bigint NOT NULL REFERENCES sport_type(sport_type_id),
	experience_years integer NOT NULL,
	sports_rank text,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	PRIMARY KEY (user_id, sport_type_id)
);

INSERT INTO sport_type (sport_type_id, name) VALUES (1, 'Running');
`

func TestRepositoriesIntegration(t *testing.T) {
	dsn := os.Getenv("PROFILE_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("PROFILE_TEST_DATABASE_DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(repositorySchemaSQL); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	profileRepository := postgres.NewProfileRepository(db)
	sportTypeRepository := postgres.NewSportTypeRepository(db)
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)

	careerSinceDate := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	err = profileRepository.Create(context.Background(), domain.Profile{
		UserID:    7,
		Username:  "coach_john",
		FirstName: "John",
		LastName:  "Doe",
		IsTrainer: true,
		CreatedAt: now,
		UpdatedAt: now,
		TrainerDetails: &domain.TrainerDetails{
			CareerSinceDate: &careerSinceDate,
			Sports: []domain.TrainerSport{{
				SportTypeID:     1,
				ExperienceYears: 5,
			}},
		},
	})
	if err != nil {
		t.Fatalf("create profile: %v", err)
	}

	profile, err := profileRepository.GetByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	if profile.Username != "coach_john" || profile.TrainerDetails == nil || len(profile.TrainerDetails.Sports) != 1 {
		t.Fatalf("unexpected profile: %+v", profile)
	}

	sportTypes, err := sportTypeRepository.ListSportTypes(context.Background())
	if err != nil {
		t.Fatalf("list sport types: %v", err)
	}
	if len(sportTypes) != 1 || sportTypes[0].ID != 1 {
		t.Fatalf("unexpected sport types: %+v", sportTypes)
	}

	authors, err := profileRepository.SearchAuthors(context.Background(), usecase.SearchAuthorsQuery{
		Query: "coach",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("search authors: %v", err)
	}
	if len(authors) != 1 || authors[0].UserID != 7 {
		t.Fatalf("unexpected authors: %+v", authors)
	}

	minExperienceYears := int32(5)
	maxExperienceYears := int32(5)
	filteredAuthors, err := profileRepository.SearchAuthors(context.Background(), usecase.SearchAuthorsQuery{
		SportTypeIDs:       []int64{1},
		MinExperienceYears: &minExperienceYears,
		MaxExperienceYears: &maxExperienceYears,
		Limit:              10,
	})
	if err != nil {
		t.Fatalf("search authors with filters: %v", err)
	}
	if len(filteredAuthors) != 1 || filteredAuthors[0].UserID != 7 {
		t.Fatalf("unexpected filtered authors: %+v", filteredAuthors)
	}
}
