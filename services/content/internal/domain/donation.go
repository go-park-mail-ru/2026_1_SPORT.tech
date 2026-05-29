package domain

import "time"

type Donation struct {
	DonationID      int64
	SenderUserID    int64
	RecipientUserID int64
	AmountValue     int32
	Currency        string
	Message         *string
	CreatedAt       time.Time
}

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusConfirmed PaymentStatus = "confirmed"
	PaymentStatusCanceled  PaymentStatus = "canceled"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusExpired   PaymentStatus = "expired"
)

func (status PaymentStatus) IsTerminal() bool {
	switch status {
	case PaymentStatusConfirmed, PaymentStatusCanceled, PaymentStatusFailed, PaymentStatusExpired:
		return true
	default:
		return false
	}
}

type DonationPayment struct {
	PaymentID         int64
	Provider          string
	ProviderPaymentID string
	Status            PaymentStatus
	SenderUserID      int64
	RecipientUserID   int64
	AmountValue       int32
	Currency          string
	Message           *string
	ConfirmationToken string
	ConfirmationURL   string
	TierID            *int64
	Donation          *Donation
	Subscription      *Subscription
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ConfirmedAt       *time.Time
}

type Balance struct {
	TrainerUserID int64
	AmountValue   int32
	Currency      string
}

type TrainerStatistics struct {
	TrainerUserID  int64
	PostsCount     int32
	DonationsCount int32
	TotalRevenue   int32
	MonthlyRevenue int32
	Currency       string
}
