package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

type SportTypeRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewSportTypeRepository(db *sql.DB, logger *slog.Logger) *SportTypeRepository {
	return &SportTypeRepository{
		db:     db,
		logger: logger,
	}
}

func (repository *SportTypeRepository) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	const query = `
		SELECT sport_type_id, name
		FROM sport_type
		ORDER BY sport_type_id
	`

	rows, err := queryContext(ctx, repository.db, repository.logger, "sport_type.list", query)
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
