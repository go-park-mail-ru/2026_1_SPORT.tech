package domain

import "time"

type Donation struct {
	DonationID      int64
	SenderUserID    int64
	RecipientUserID int64
	AmountValue     int64
	Currency        string
	Message         *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
