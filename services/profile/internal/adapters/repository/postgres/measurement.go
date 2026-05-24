package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
)

type MeasurementRepository struct {
	db *sql.DB
}

func NewMeasurementRepository(db *sql.DB) *MeasurementRepository {
	return &MeasurementRepository{db: db}
}

func (r *MeasurementRepository) CreateMeasurement(ctx context.Context, m domain.Measurement) (domain.Measurement, error) {
	const query = `
		INSERT INTO measurement (
			user_id, measured_at, weight_kg, body_fat_pct, chest_cm, waist_cm, hips_cm, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING measurement_id, created_at, updated_at
	`
	now := time.Now().UTC()
	row := r.db.QueryRowContext(ctx, query,
		m.UserID,
		m.MeasuredAt.Format("2006-01-02"),
		nullableFloat64(m.WeightKg),
		nullableFloat64(m.BodyFatPct),
		nullableInt32(m.ChestCm),
		nullableInt32(m.WaistCm),
		nullableInt32(m.HipsCm),
		nullString(m.Notes),
		now,
	)
	if err := row.Scan(&m.MeasurementID, &m.CreatedAt, &m.UpdatedAt); err != nil {
		return domain.Measurement{}, err
	}
	return m, nil
}

func (r *MeasurementRepository) ListMeasurements(ctx context.Context, userID int64, limit, offset int32) ([]domain.Measurement, error) {
	const query = `
		SELECT measurement_id, user_id, measured_at, weight_kg, body_fat_pct,
		       chest_cm, waist_cm, hips_cm, notes, created_at, updated_at
		FROM measurement
		WHERE user_id = $1
		ORDER BY measured_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Measurement
	for rows.Next() {
		var m domain.Measurement
		var weightKg, bodyFatPct sql.NullFloat64
		var chestCm, waistCm, hipsCm sql.NullInt32
		var notes sql.NullString
		if err := rows.Scan(
			&m.MeasurementID, &m.UserID, &m.MeasuredAt,
			&weightKg, &bodyFatPct, &chestCm, &waistCm, &hipsCm,
			&notes, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if weightKg.Valid {
			v := weightKg.Float64
			m.WeightKg = &v
		}
		if bodyFatPct.Valid {
			v := bodyFatPct.Float64
			m.BodyFatPct = &v
		}
		if chestCm.Valid {
			v := chestCm.Int32
			m.ChestCm = &v
		}
		if waistCm.Valid {
			v := waistCm.Int32
			m.WaistCm = &v
		}
		if hipsCm.Valid {
			v := hipsCm.Int32
			m.HipsCm = &v
		}
		if notes.Valid {
			m.Notes = &notes.String
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *MeasurementRepository) DeleteMeasurement(ctx context.Context, userID, measurementID int64) error {
	const query = `DELETE FROM measurement WHERE measurement_id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, measurementID, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return usecase.ErrMeasurementNotFound
	}
	return nil
}

func nullableFloat64(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullableInt32(v *int32) any {
	if v == nil {
		return nil
	}
	return *v
}
