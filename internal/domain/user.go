package domain

import "time"

type User struct {
	ID           int64
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	IsTrainer    bool
	IsAdmin      bool
	FirstName    string
	LastName     string
	Bio          *string
	AvatarURL    *string
}
