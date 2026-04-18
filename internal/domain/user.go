package domain

import "time"

type TrainerSport struct {
	SportTypeID     int64
	ExperienceYears int
	SportsRank      *string
}

type TrainerDetails struct {
	EducationDegree *string
	CareerSinceDate time.Time
	Sports          []TrainerSport
}

type TrainerListItem struct {
	ID             int64
	Username       string
	FirstName      string
	LastName       string
	Bio            *string
	AvatarURL      *string
	TrainerDetails *TrainerDetails
}

type User struct {
	ID             int64
	Username       string
	Email          string
	PasswordHash   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	IsTrainer      bool
	IsAdmin        bool
	FirstName      string
	LastName       string
	Bio            *string
	AvatarURL      *string
	TrainerDetails *TrainerDetails
}
