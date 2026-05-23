package domain

import "time"

// Measurement хранит замеры прогресса клиента.
type Measurement struct {
	MeasurementID int64
	UserID        int64
	MeasuredAt    time.Time
	WeightKg      *float64
	BodyFatPct    *float64
	ChestCm       *int32
	WaistCm       *int32
	HipsCm        *int32
	Notes         *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
