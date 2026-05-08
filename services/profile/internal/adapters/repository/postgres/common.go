package postgres

import (
	"errors"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/lib/pq"
	"time"
)

func nullString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullTime(value *time.Time) any {
	if value == nil {
		return nil
	}

	return *value
}

func mapProfileError(err error) error {
	var postgresError *pq.Error
	if !errors.As(err, &postgresError) {
		return err
	}

	switch postgresError.Code {
	case "23505":
		switch postgresError.Constraint {
		case "profile_pkey":
			return domain.ErrProfileExists
		case "profile_username_key":
			return domain.ErrUsernameTaken
		default:
			return err
		}
	case "23503":
		switch postgresError.Constraint {
		case "trainer_sport_sport_type_id_fkey":
			return domain.ErrSportTypeNotFound
		default:
			return err
		}
	default:
		return err
	}
}
