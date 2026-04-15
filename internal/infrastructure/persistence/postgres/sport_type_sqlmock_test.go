package postgres

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSportTypeRepositoryListSportTypes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repository := NewSportTypeRepository(db, nil)

	rows := sqlmock.NewRows([]string{"sport_type_id", "name"}).
		AddRow(int64(1), "Бег").
		AddRow(int64(2), "Плавание")

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT sport_type_id, name
		FROM sport_type
		ORDER BY sport_type_id
	`)).
		WillReturnRows(rows)

	sportTypes, err := repository.ListSportTypes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sportTypes) != 2 {
		t.Fatalf("unexpected sport types: %+v", sportTypes)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
