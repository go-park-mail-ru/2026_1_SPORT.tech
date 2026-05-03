package domain

import "time"

type Subscription struct {
	SubscriptionID int64
	ClientUserID   int64
	TrainerUserID  int64
	TierID         int64
	TierName       string
	Price          int32
	Active         bool
	ExpiresAt      time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
