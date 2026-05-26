package domain

import "time"

type MeetingBookingStatus string

const (
	MeetingBookingStatusConfirmed MeetingBookingStatus = "confirmed"
	MeetingBookingStatusCancelled MeetingBookingStatus = "cancelled"
)

type MeetingAvailabilityRule struct {
	RuleID        int64
	TrainerUserID int64
	Weekday       int32
	StartHour     int32
	CreatedAt     time.Time
}

type MeetingSlot struct {
	SlotID        int64
	TrainerUserID int64
	StartsAt      time.Time
	CreatedAt     time.Time
}

type MeetingAvailabilitySlot struct {
	StartsAt time.Time
	EndsAt   time.Time
}

type MeetingBooking struct {
	BookingID         int64
	TrainerUserID     int64
	ClientUserID      int64
	StartsAt          time.Time
	EndsAt            time.Time
	Status            MeetingBookingStatus
	CreatedByUserID   int64
	Note              *string
	CreatedAt         time.Time
	CancelledAt       *time.Time
	CancelledByUserID *int64
}
