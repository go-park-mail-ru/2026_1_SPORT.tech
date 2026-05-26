package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/lib/pq"
)

func (repository *Repository) HasActiveCalendarSubscription(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
	var exists bool
	err := repository.db.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM content_subscription cs
				JOIN content_subscription_tier cst
					ON cst.trainer_user_id = cs.trainer_user_id
					AND cst.tier_id = cs.tier_id
				WHERE cs.client_user_id = $1
					AND cs.trainer_user_id = $2
					AND cs.active = TRUE
					AND cs.expires_at > now()
					AND cst.calendar_enabled = TRUE
			)
		`,
		clientUserID,
		trainerUserID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (repository *Repository) CreateMeetingAvailabilityRule(ctx context.Context, rule domain.MeetingAvailabilityRule) (domain.MeetingAvailabilityRule, error) {
	row := repository.db.QueryRowContext(
		ctx,
		`
			INSERT INTO content_meeting_availability_rule (trainer_user_id, weekday, start_hour)
			VALUES ($1, $2, $3)
			ON CONFLICT (trainer_user_id, weekday, start_hour) DO UPDATE
				SET trainer_user_id = EXCLUDED.trainer_user_id
			RETURNING rule_id, trainer_user_id, weekday, start_hour, created_at
		`,
		rule.TrainerUserID,
		rule.Weekday,
		rule.StartHour,
	)

	return scanMeetingRule(row)
}

func (repository *Repository) ListMeetingAvailabilityRules(ctx context.Context, trainerUserID int64) ([]domain.MeetingAvailabilityRule, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
			SELECT rule_id, trainer_user_id, weekday, start_hour, created_at
			FROM content_meeting_availability_rule
			WHERE trainer_user_id = $1
			ORDER BY weekday ASC, start_hour ASC
		`,
		trainerUserID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]domain.MeetingAvailabilityRule, 0)
	for rows.Next() {
		rule, err := scanMeetingRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

func (repository *Repository) DeleteMeetingAvailabilityRule(ctx context.Context, trainerUserID int64, ruleID int64) error {
	result, err := repository.db.ExecContext(
		ctx,
		`
			DELETE FROM content_meeting_availability_rule
			WHERE rule_id = $1 AND trainer_user_id = $2
		`,
		ruleID,
		trainerUserID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrMeetingRuleNotFound
	}

	return nil
}

func (repository *Repository) CreateMeetingSlot(ctx context.Context, slot domain.MeetingSlot) (domain.MeetingSlot, error) {
	row := repository.db.QueryRowContext(
		ctx,
		`
			INSERT INTO content_meeting_slot (trainer_user_id, starts_at)
			VALUES ($1, $2)
			ON CONFLICT (trainer_user_id, starts_at) DO UPDATE
				SET trainer_user_id = EXCLUDED.trainer_user_id
			RETURNING slot_id, trainer_user_id, starts_at, created_at
		`,
		slot.TrainerUserID,
		slot.StartsAt.UTC(),
	)

	return scanMeetingSlot(row)
}

func (repository *Repository) DeleteMeetingSlot(ctx context.Context, trainerUserID int64, slotID int64) error {
	result, err := repository.db.ExecContext(
		ctx,
		`
			DELETE FROM content_meeting_slot
			WHERE slot_id = $1 AND trainer_user_id = $2
		`,
		slotID,
		trainerUserID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrMeetingSlotNotFound
	}

	return nil
}

func (repository *Repository) ListMeetingSlots(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingSlot, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
			SELECT slot_id, trainer_user_id, starts_at, created_at
			FROM content_meeting_slot
			WHERE trainer_user_id = $1
				AND starts_at >= $2
				AND starts_at < $3
			ORDER BY starts_at ASC
		`,
		trainerUserID,
		from.UTC(),
		to.UTC(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := make([]domain.MeetingSlot, 0)
	for rows.Next() {
		slot, err := scanMeetingSlot(rows)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	return slots, rows.Err()
}

func (repository *Repository) ListConfirmedBookings(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingBooking, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
			SELECT booking_id, trainer_user_id, client_user_id, starts_at, ends_at,
				status, created_by_user_id, note, created_at, cancelled_at, cancelled_by_user_id
			FROM content_meeting_booking
			WHERE trainer_user_id = $1
				AND status = 'confirmed'
				AND starts_at < $3
				AND ends_at > $2
			ORDER BY starts_at ASC
		`,
		trainerUserID,
		from.UTC(),
		to.UTC(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMeetingBookings(rows)
}

func (repository *Repository) CountActiveClientBookings(ctx context.Context, clientUserID int64, trainerUserID int64, now time.Time) (int32, error) {
	var count int32
	err := repository.db.QueryRowContext(
		ctx,
		`
			SELECT COUNT(*)
			FROM content_meeting_booking
			WHERE client_user_id = $1
				AND trainer_user_id = $2
				AND status = 'confirmed'
				AND starts_at > $3
		`,
		clientUserID,
		trainerUserID,
		now.UTC(),
	).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repository *Repository) CreateBooking(ctx context.Context, booking domain.MeetingBooking) (domain.MeetingBooking, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.MeetingBooking{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var conflict bool
	if err := tx.QueryRowContext(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM content_meeting_booking
				WHERE trainer_user_id = $1
					AND status = 'confirmed'
					AND starts_at < $3
					AND ends_at > $2
			)
		`,
		booking.TrainerUserID,
		booking.StartsAt.UTC(),
		booking.EndsAt.UTC(),
	).Scan(&conflict); err != nil {
		return domain.MeetingBooking{}, err
	}
	if conflict {
		return domain.MeetingBooking{}, domain.ErrMeetingSlotTaken
	}

	row := tx.QueryRowContext(
		ctx,
		`
			INSERT INTO content_meeting_booking (
				trainer_user_id, client_user_id, starts_at, ends_at, status, created_by_user_id, note
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING booking_id, trainer_user_id, client_user_id, starts_at, ends_at,
				status, created_by_user_id, note, created_at, cancelled_at, cancelled_by_user_id
		`,
		booking.TrainerUserID,
		booking.ClientUserID,
		booking.StartsAt.UTC(),
		booking.EndsAt.UTC(),
		string(booking.Status),
		booking.CreatedByUserID,
		nullString(booking.Note),
	)

	created, err := scanMeetingBooking(row)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.MeetingBooking{}, domain.ErrMeetingSlotTaken
		}
		return domain.MeetingBooking{}, err
	}

	if err := tx.Commit(); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.MeetingBooking{}, domain.ErrMeetingSlotTaken
		}
		return domain.MeetingBooking{}, err
	}

	return created, nil
}

func (repository *Repository) GetBooking(ctx context.Context, bookingID int64) (domain.MeetingBooking, error) {
	row := repository.db.QueryRowContext(
		ctx,
		`
			SELECT booking_id, trainer_user_id, client_user_id, starts_at, ends_at,
				status, created_by_user_id, note, created_at, cancelled_at, cancelled_by_user_id
			FROM content_meeting_booking
			WHERE booking_id = $1
		`,
		bookingID,
	)

	booking, err := scanMeetingBooking(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.MeetingBooking{}, domain.ErrMeetingBookingNotFound
		}
		return domain.MeetingBooking{}, err
	}

	return booking, nil
}

func (repository *Repository) CancelBooking(ctx context.Context, bookingID int64, cancelledByUserID int64) (domain.MeetingBooking, error) {
	row := repository.db.QueryRowContext(
		ctx,
		`
			UPDATE content_meeting_booking
			SET status = 'cancelled',
				cancelled_at = now(),
				cancelled_by_user_id = $2
			WHERE booking_id = $1
				AND status = 'confirmed'
			RETURNING booking_id, trainer_user_id, client_user_id, starts_at, ends_at,
				status, created_by_user_id, note, created_at, cancelled_at, cancelled_by_user_id
		`,
		bookingID,
		cancelledByUserID,
	)

	booking, err := scanMeetingBooking(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.MeetingBooking{}, domain.ErrMeetingBookingNotFound
		}
		return domain.MeetingBooking{}, err
	}

	return booking, nil
}

func (repository *Repository) ListUserBookings(ctx context.Context, userID int64) ([]domain.MeetingBooking, error) {
	rows, err := repository.db.QueryContext(
		ctx,
		`
			SELECT booking_id, trainer_user_id, client_user_id, starts_at, ends_at,
				status, created_by_user_id, note, created_at, cancelled_at, cancelled_by_user_id
			FROM content_meeting_booking
			WHERE trainer_user_id = $1 OR client_user_id = $1
			ORDER BY starts_at DESC
		`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMeetingBookings(rows)
}

func scanMeetingRule(scanner sqlScanner) (domain.MeetingAvailabilityRule, error) {
	var rule domain.MeetingAvailabilityRule
	if err := scanner.Scan(
		&rule.RuleID,
		&rule.TrainerUserID,
		&rule.Weekday,
		&rule.StartHour,
		&rule.CreatedAt,
	); err != nil {
		return domain.MeetingAvailabilityRule{}, err
	}

	return rule, nil
}

func scanMeetingSlot(scanner sqlScanner) (domain.MeetingSlot, error) {
	var slot domain.MeetingSlot
	if err := scanner.Scan(
		&slot.SlotID,
		&slot.TrainerUserID,
		&slot.StartsAt,
		&slot.CreatedAt,
	); err != nil {
		return domain.MeetingSlot{}, err
	}

	return slot, nil
}

func scanMeetingBooking(scanner sqlScanner) (domain.MeetingBooking, error) {
	var (
		booking         domain.MeetingBooking
		status          string
		note            sql.NullString
		cancelledAt     sql.NullTime
		cancelledByUser sql.NullInt64
	)
	if err := scanner.Scan(
		&booking.BookingID,
		&booking.TrainerUserID,
		&booking.ClientUserID,
		&booking.StartsAt,
		&booking.EndsAt,
		&status,
		&booking.CreatedByUserID,
		&note,
		&booking.CreatedAt,
		&cancelledAt,
		&cancelledByUser,
	); err != nil {
		return domain.MeetingBooking{}, err
	}

	booking.Status = domain.MeetingBookingStatus(status)
	if note.Valid {
		booking.Note = &note.String
	}
	booking.CancelledAt = timePtrFromNull(cancelledAt)
	if cancelledByUser.Valid {
		booking.CancelledByUserID = &cancelledByUser.Int64
	}

	return booking, nil
}

func scanMeetingBookings(rows *sql.Rows) ([]domain.MeetingBooking, error) {
	bookings := make([]domain.MeetingBooking, 0)
	for rows.Next() {
		booking, err := scanMeetingBooking(rows)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}
