package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

// ── stub MeasurementRepository ────────────────────────────────────────────────

type stubMeasurementRepository struct {
	createFunc func(ctx context.Context, m domain.Measurement) (domain.Measurement, error)
	listFunc   func(ctx context.Context, userID int64, limit, offset int32) ([]domain.Measurement, error)
	deleteFunc func(ctx context.Context, userID, measurementID int64) error
}

func (r stubMeasurementRepository) CreateMeasurement(ctx context.Context, m domain.Measurement) (domain.Measurement, error) {
	return r.createFunc(ctx, m)
}

func (r stubMeasurementRepository) ListMeasurements(ctx context.Context, userID int64, limit, offset int32) ([]domain.Measurement, error) {
	return r.listFunc(ctx, userID, limit, offset)
}

func (r stubMeasurementRepository) DeleteMeasurement(ctx context.Context, userID, measurementID int64) error {
	return r.deleteFunc(ctx, userID, measurementID)
}

func newMeasurementService(stub stubMeasurementRepository) *Service {
	return &Service{
		measurements: stub,
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func floatPtr(f float64) *float64 { return &f }
func int32Ptr(i int32) *int32     { return &i }
func strPtr(s string) *string     { return &s }

// ── CreateMeasurement ─────────────────────────────────────────────────────────

func TestCreateMeasurement_Success(t *testing.T) {
	t.Parallel()
	now := time.Now()
	stub := stubMeasurementRepository{
		createFunc: func(_ context.Context, m domain.Measurement) (domain.Measurement, error) {
			m.MeasurementID = 1
			m.CreatedAt = now
			m.UpdatedAt = now
			return m, nil
		},
	}
	svc := newMeasurementService(stub)
	cmd := CreateMeasurementCommand{
		UserID:     1,
		MeasuredAt: "2025-01-15",
		WeightKg:   floatPtr(80.5),
	}
	got, err := svc.CreateMeasurement(context.Background(), cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MeasurementID != 1 {
		t.Fatalf("expected measurement_id=1, got %d", got.MeasurementID)
	}
	if *got.WeightKg != 80.5 {
		t.Fatalf("expected weight_kg=80.5, got %v", *got.WeightKg)
	}
}

func TestCreateMeasurement_InvalidUserID(t *testing.T) {
	t.Parallel()
	svc := newMeasurementService(stubMeasurementRepository{})
	_, err := svc.CreateMeasurement(context.Background(), CreateMeasurementCommand{
		UserID:     0,
		MeasuredAt: "2025-01-15",
		WeightKg:   floatPtr(70),
	})
	if err != ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestCreateMeasurement_InvalidDate(t *testing.T) {
	t.Parallel()
	svc := newMeasurementService(stubMeasurementRepository{})
	_, err := svc.CreateMeasurement(context.Background(), CreateMeasurementCommand{
		UserID:     1,
		MeasuredAt: "not-a-date",
		WeightKg:   floatPtr(70),
	})
	if err != ErrInvalidMeasuredAt {
		t.Fatalf("expected ErrInvalidMeasuredAt, got %v", err)
	}
}

func TestCreateMeasurement_NoValues(t *testing.T) {
	t.Parallel()
	svc := newMeasurementService(stubMeasurementRepository{})
	_, err := svc.CreateMeasurement(context.Background(), CreateMeasurementCommand{
		UserID:     1,
		MeasuredAt: "2025-01-15",
	})
	if err != ErrInvalidMeasurementData {
		t.Fatalf("expected ErrInvalidMeasurementData, got %v", err)
	}
}

func TestCreateMeasurement_AllFields(t *testing.T) {
	t.Parallel()
	stub := stubMeasurementRepository{
		createFunc: func(_ context.Context, m domain.Measurement) (domain.Measurement, error) {
			m.MeasurementID = 7
			return m, nil
		},
	}
	svc := newMeasurementService(stub)
	cmd := CreateMeasurementCommand{
		UserID:     2,
		MeasuredAt: "2025-06-01",
		WeightKg:   floatPtr(75.0),
		BodyFatPct: floatPtr(18.5),
		ChestCm:    int32Ptr(100),
		WaistCm:    int32Ptr(78),
		HipsCm:     int32Ptr(95),
		Notes:      strPtr("Отличная неделя"),
	}
	got, err := svc.CreateMeasurement(context.Background(), cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MeasurementID != 7 {
		t.Fatalf("expected id=7")
	}
}

// ── ListMeasurements ──────────────────────────────────────────────────────────

func TestListMeasurements_Success(t *testing.T) {
	t.Parallel()
	now := time.Now()
	expected := []domain.Measurement{
		{MeasurementID: 1, UserID: 1, MeasuredAt: now, WeightKg: floatPtr(80)},
		{MeasurementID: 2, UserID: 1, MeasuredAt: now.AddDate(0, 0, -7), WeightKg: floatPtr(81)},
	}
	stub := stubMeasurementRepository{
		listFunc: func(_ context.Context, userID int64, limit, offset int32) ([]domain.Measurement, error) {
			return expected, nil
		},
	}
	svc := newMeasurementService(stub)
	got, err := svc.ListMeasurements(context.Background(), ListMeasurementsQuery{UserID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 measurements, got %d", len(got))
	}
}

func TestListMeasurements_InvalidUserID(t *testing.T) {
	t.Parallel()
	svc := newMeasurementService(stubMeasurementRepository{})
	_, err := svc.ListMeasurements(context.Background(), ListMeasurementsQuery{UserID: -1})
	if err != ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestListMeasurements_DefaultLimit(t *testing.T) {
	t.Parallel()
	var gotLimit int32
	stub := stubMeasurementRepository{
		listFunc: func(_ context.Context, _ int64, limit, _ int32) ([]domain.Measurement, error) {
			gotLimit = limit
			return nil, nil
		},
	}
	svc := newMeasurementService(stub)
	if _, err := svc.ListMeasurements(context.Background(), ListMeasurementsQuery{UserID: 1, Limit: 0}); err != nil {
		t.Fatal(err)
	}
	if gotLimit != 100 {
		t.Fatalf("expected default limit 100, got %d", gotLimit)
	}
}

// ── DeleteMeasurement ─────────────────────────────────────────────────────────

func TestDeleteMeasurement_Success(t *testing.T) {
	t.Parallel()
	stub := stubMeasurementRepository{
		deleteFunc: func(_ context.Context, _, _ int64) error { return nil },
	}
	svc := newMeasurementService(stub)
	if err := svc.DeleteMeasurement(context.Background(), DeleteMeasurementCommand{UserID: 1, MeasurementID: 5}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMeasurement_InvalidUserID(t *testing.T) {
	t.Parallel()
	svc := newMeasurementService(stubMeasurementRepository{})
	err := svc.DeleteMeasurement(context.Background(), DeleteMeasurementCommand{UserID: 0, MeasurementID: 1})
	if err != ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestDeleteMeasurement_NotFound(t *testing.T) {
	t.Parallel()
	stub := stubMeasurementRepository{
		deleteFunc: func(_ context.Context, _, _ int64) error { return ErrMeasurementNotFound },
	}
	svc := newMeasurementService(stub)
	err := svc.DeleteMeasurement(context.Background(), DeleteMeasurementCommand{UserID: 1, MeasurementID: 99})
	if err != ErrMeasurementNotFound {
		t.Fatalf("expected ErrMeasurementNotFound, got %v", err)
	}
}
