package postgres

import (
	"context"
	"database/sql"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

type SportTypeRepository struct {
	db *sql.DB
}

func NewSportTypeRepository(db *sql.DB) *SportTypeRepository {
	return &SportTypeRepository{
		db: db,
	}
}

func (repository *SportTypeRepository) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	const query = `
		SELECT sport_type_id, name
		FROM sport_type
		ORDER BY sport_type_id
	`

	rows, err := repository.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sportTypes := make([]domain.SportType, 0)
	for rows.Next() {
		var sportType domain.SportType
		if err := rows.Scan(&sportType.ID, &sportType.Name); err != nil {
			return nil, err
		}

		sportTypes = append(sportTypes, sportType)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sportTypes, nil
}
