package domain

import "time"

type TrainerSport struct {
	SportTypeID     int64
	ExperienceYears int
	SportsRank      *string
}

type TrainerDetails struct {
	EducationDegree *string
	CareerSinceDate *time.Time
	Sports          []TrainerSport
}

type Profile struct {
	UserID         int64
	Username       string
	FirstName      string
	LastName       string
	Bio            *string
	AvatarURL      *string
	IsTrainer      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	TrainerDetails *TrainerDetails
}

type AuthorSummary struct {
	UserID         int64
	Username       string
	FirstName      string
	LastName       string
	Bio            *string
	AvatarURL      *string
	TrainerDetails *TrainerDetails
}

func (profile Profile) EnsureTrainer() error {
	if !profile.IsTrainer {
		return ErrTrainerProfileForbidden
	}

	return nil
}
