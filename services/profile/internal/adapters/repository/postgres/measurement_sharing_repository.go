package postgres

import (
	"context"
	"database/sql"
	"time"
)

// MeasurementSharingRepository хранит информацию о том, каким тренерам
// клиент разрешил видеть свои замеры.
type MeasurementSharingRepository struct {
	db *sql.DB
}

func NewMeasurementSharingRepository(db *sql.DB) *MeasurementSharingRepository {
	return &MeasurementSharingRepository{db: db}
}

// SetSharing заменяет текущий список разрешённых тренеров для clientUserID
// на переданный trainerUserIDs. Операция атомарна (DELETE + INSERT в транзакции).
func (r *MeasurementSharingRepository) SetSharing(ctx context.Context, clientUserID int64, trainerUserIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM measurement_sharing WHERE client_user_id = $1`, clientUserID,
	); err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, trainerID := range trainerUserIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO measurement_sharing (client_user_id, trainer_user_id, created_at)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (client_user_id, trainer_user_id) DO NOTHING`,
			clientUserID, trainerID, now,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetSharing возвращает список trainer_user_id, которым clientUserID разрешил доступ.
func (r *MeasurementSharingRepository) GetSharing(ctx context.Context, clientUserID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT trainer_user_id FROM measurement_sharing WHERE client_user_id = $1 ORDER BY created_at`,
		clientUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// HasAccess проверяет, разрешил ли clientUserID просматривать свои замеры трейнеру trainerUserID.
func (r *MeasurementSharingRepository) HasAccess(ctx context.Context, clientUserID, trainerUserID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM measurement_sharing
			WHERE client_user_id = $1 AND trainer_user_id = $2
		)`,
		clientUserID, trainerUserID,
	).Scan(&exists)
	return exists, err
}
