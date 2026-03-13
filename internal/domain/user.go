package domain

import "time"

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
