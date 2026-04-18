package domain

import "time"

type Session struct {
	IDHash    string
	UserID    int64
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (session Session) IsExpired(now time.Time) bool {
	return !now.Before(session.ExpiresAt)
}

func (session Session) IsActive(now time.Time) bool {
	return session.RevokedAt == nil && !session.IsExpired(now)
}
