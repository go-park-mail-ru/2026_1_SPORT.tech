package domain

import "time"

type Session struct {
	SessionIDHash string
	UserID        int64
	ExpiresAt     time.Time
}
