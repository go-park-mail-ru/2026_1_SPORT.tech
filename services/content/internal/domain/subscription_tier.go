package domain

import "time"

type SubscriptionTier struct {
	TierID        int64
	TrainerUserID int64
	Name          string
	Price         int32
	Description   *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
