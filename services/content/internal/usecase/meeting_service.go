package usecase

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

const (
	meetingSlotDuration      = time.Hour
	maxMeetingDurationHours  = 8
	defaultAvailabilityDays  = 14
	maxAvailabilityRangeDays = 31
	maxMeetingNoteLength     = 1000
)

func (service *Service) CreateMeetingAvailabilityRule(ctx context.Context, command CreateMeetingAvailabilityRuleCommand) (domain.MeetingAvailabilityRule, error) {
	if command.TrainerUserID <= 0 {
		return domain.MeetingAvailabilityRule{}, ErrInvalidUserID
	}
	if command.Weekday < 0 || command.Weekday > 6 {
		return domain.MeetingAvailabilityRule{}, ErrInvalidMeetingWeekday
	}
	if command.StartHour < 0 || command.StartHour > 23 {
		return domain.MeetingAvailabilityRule{}, ErrInvalidMeetingHour
	}

	return service.meeting.CreateMeetingAvailabilityRule(ctx, domain.MeetingAvailabilityRule{
		TrainerUserID: command.TrainerUserID,
		Weekday:       command.Weekday,
		StartHour:     command.StartHour,
	})
}

func (service *Service) ListMeetingAvailabilityRules(ctx context.Context, query ListMeetingAvailabilityRulesQuery) ([]domain.MeetingAvailabilityRule, error) {
	if query.TrainerUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	return service.meeting.ListMeetingAvailabilityRules(ctx, query.TrainerUserID)
}

func (service *Service) DeleteMeetingAvailabilityRule(ctx context.Context, command DeleteMeetingAvailabilityRuleCommand) error {
	if command.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.RuleID <= 0 {
		return ErrInvalidMeetingRuleID
	}

	return service.meeting.DeleteMeetingAvailabilityRule(ctx, command.TrainerUserID, command.RuleID)
}

func (service *Service) CreateMeetingSlot(ctx context.Context, command CreateMeetingSlotCommand) (domain.MeetingSlot, error) {
	if command.TrainerUserID <= 0 {
		return domain.MeetingSlot{}, ErrInvalidUserID
	}
	startsAt, err := normalizeMeetingStart(command.StartsAt)
	if err != nil {
		return domain.MeetingSlot{}, err
	}

	return service.meeting.CreateMeetingSlot(ctx, domain.MeetingSlot{
		TrainerUserID: command.TrainerUserID,
		StartsAt:      startsAt,
	})
}

func (service *Service) DeleteMeetingSlot(ctx context.Context, command DeleteMeetingSlotCommand) error {
	if command.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.SlotID <= 0 {
		return ErrInvalidMeetingSlotID
	}

	return service.meeting.DeleteMeetingSlot(ctx, command.TrainerUserID, command.SlotID)
}

func (service *Service) ListTrainerMeetingAvailability(ctx context.Context, query ListTrainerMeetingAvailabilityQuery) ([]domain.MeetingAvailabilitySlot, error) {
	if query.TrainerUserID <= 0 || query.ViewerUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	if query.ViewerUserID != query.TrainerUserID {
		allowed, err := service.meeting.HasActiveCalendarSubscription(ctx, query.ViewerUserID, query.TrainerUserID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, domain.ErrMeetingAccessForbidden
		}
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	from, err := parseMeetingDate(query.From, today)
	if err != nil {
		return nil, err
	}
	to, err := parseMeetingDate(query.To, from.AddDate(0, 0, defaultAvailabilityDays))
	if err != nil {
		return nil, err
	}
	if to.Before(from) {
		return nil, ErrInvalidMeetingRange
	}
	if to.Sub(from) > maxAvailabilityRangeDays*24*time.Hour {
		to = from.AddDate(0, 0, maxAvailabilityRangeDays)
	}

	rangeStart := from
	rangeEnd := to.AddDate(0, 0, 1)

	rules, err := service.meeting.ListMeetingAvailabilityRules(ctx, query.TrainerUserID)
	if err != nil {
		return nil, err
	}
	slots, err := service.meeting.ListMeetingSlots(ctx, query.TrainerUserID, rangeStart, rangeEnd)
	if err != nil {
		return nil, err
	}
	bookings, err := service.meeting.ListConfirmedBookings(ctx, query.TrainerUserID, rangeStart, rangeEnd)
	if err != nil {
		return nil, err
	}

	candidates := make(map[int64]time.Time)
	for day := rangeStart; day.Before(rangeEnd); day = day.AddDate(0, 0, 1) {
		weekday := int32(day.Weekday())
		for _, rule := range rules {
			if rule.Weekday != weekday {
				continue
			}
			candidate := time.Date(day.Year(), day.Month(), day.Day(), int(rule.StartHour), 0, 0, 0, time.UTC)
			candidates[candidate.Unix()] = candidate
		}
	}
	for _, slot := range slots {
		candidate := slot.StartsAt.UTC().Truncate(time.Hour)
		candidates[candidate.Unix()] = candidate
	}

	available := make([]domain.MeetingAvailabilitySlot, 0, len(candidates))
	for _, candidate := range candidates {
		if !candidate.After(now) {
			continue
		}
		end := candidate.Add(meetingSlotDuration)
		if overlapsAnyBooking(candidate, end, bookings) {
			continue
		}
		available = append(available, domain.MeetingAvailabilitySlot{StartsAt: candidate, EndsAt: end})
	}

	sort.Slice(available, func(i, j int) bool {
		return available[i].StartsAt.Before(available[j].StartsAt)
	})

	return available, nil
}

func (service *Service) BookMeeting(ctx context.Context, command BookMeetingCommand) (domain.MeetingBooking, error) {
	if command.ClientUserID <= 0 || command.TrainerUserID <= 0 {
		return domain.MeetingBooking{}, ErrInvalidUserID
	}
	if command.ClientUserID == command.TrainerUserID {
		return domain.MeetingBooking{}, ErrInvalidUserID
	}
	startsAt, err := normalizeMeetingStart(command.StartsAt)
	if err != nil {
		return domain.MeetingBooking{}, err
	}

	allowed, err := service.meeting.HasActiveCalendarSubscription(ctx, command.ClientUserID, command.TrainerUserID)
	if err != nil {
		return domain.MeetingBooking{}, err
	}
	if !allowed {
		return domain.MeetingBooking{}, domain.ErrMeetingAccessForbidden
	}

	offered, err := service.isMeetingSlotOffered(ctx, command.TrainerUserID, startsAt)
	if err != nil {
		return domain.MeetingBooking{}, err
	}
	if !offered {
		return domain.MeetingBooking{}, domain.ErrMeetingSlotUnavailable
	}

	activeCount, err := service.meeting.CountActiveClientBookings(ctx, command.ClientUserID, command.TrainerUserID, time.Now().UTC())
	if err != nil {
		return domain.MeetingBooking{}, err
	}
	if activeCount >= 1 {
		return domain.MeetingBooking{}, domain.ErrMeetingClientLimit
	}

	booking, err := service.meeting.CreateBooking(ctx, domain.MeetingBooking{
		TrainerUserID:   command.TrainerUserID,
		ClientUserID:    command.ClientUserID,
		StartsAt:        startsAt,
		EndsAt:          startsAt.Add(meetingSlotDuration),
		Status:          domain.MeetingBookingStatusConfirmed,
		CreatedByUserID: command.ClientUserID,
	})
	if err != nil {
		return domain.MeetingBooking{}, err
	}

	_ = service.createNotification(ctx, domain.Notification{
		UserID:      booking.TrainerUserID,
		Type:        domain.NotificationTypeMeeting,
		ActorUserID: booking.ClientUserID,
		Title:       "Новая запись",
		Body:        "Клиент записался к вам на занятие",
	})

	return booking, nil
}

func (service *Service) AssignMeeting(ctx context.Context, command AssignMeetingCommand) (domain.MeetingBooking, error) {
	if command.TrainerUserID <= 0 || command.ClientUserID <= 0 {
		return domain.MeetingBooking{}, ErrInvalidUserID
	}
	if command.TrainerUserID == command.ClientUserID {
		return domain.MeetingBooking{}, ErrInvalidUserID
	}
	if command.DurationHours < 1 || command.DurationHours > maxMeetingDurationHours {
		return domain.MeetingBooking{}, ErrInvalidMeetingDuration
	}
	startsAt, err := normalizeMeetingStart(command.StartsAt)
	if err != nil {
		return domain.MeetingBooking{}, err
	}
	note := normalizeOptionalText(command.Note)
	if note != nil && len(*note) > maxMeetingNoteLength {
		return domain.MeetingBooking{}, ErrInvalidMeetingNote
	}

	allowed, err := service.meeting.HasActiveCalendarSubscription(ctx, command.ClientUserID, command.TrainerUserID)
	if err != nil {
		return domain.MeetingBooking{}, err
	}
	if !allowed {
		return domain.MeetingBooking{}, domain.ErrMeetingAccessForbidden
	}

	booking, err := service.meeting.CreateBooking(ctx, domain.MeetingBooking{
		TrainerUserID:   command.TrainerUserID,
		ClientUserID:    command.ClientUserID,
		StartsAt:        startsAt,
		EndsAt:          startsAt.Add(time.Duration(command.DurationHours) * time.Hour),
		Status:          domain.MeetingBookingStatusConfirmed,
		CreatedByUserID: command.TrainerUserID,
		Note:            note,
	})
	if err != nil {
		return domain.MeetingBooking{}, err
	}

	_ = service.createNotification(ctx, domain.Notification{
		UserID:      booking.ClientUserID,
		Type:        domain.NotificationTypeMeeting,
		ActorUserID: booking.TrainerUserID,
		Title:       "Вам назначено занятие",
		Body:        "Тренер назначил вам занятие в календаре",
	})

	return booking, nil
}

func (service *Service) CancelMeeting(ctx context.Context, command CancelMeetingCommand) error {
	if command.UserID <= 0 {
		return ErrInvalidUserID
	}
	if command.BookingID <= 0 {
		return ErrInvalidMeetingBookingID
	}

	booking, err := service.meeting.GetBooking(ctx, command.BookingID)
	if err != nil {
		return err
	}
	if booking.TrainerUserID != command.UserID && booking.ClientUserID != command.UserID {
		return domain.ErrMeetingBookingNotFound
	}
	if booking.Status != domain.MeetingBookingStatusConfirmed {
		return domain.ErrMeetingNotCancelable
	}
	if !booking.StartsAt.After(time.Now().UTC()) {
		return domain.ErrMeetingNotCancelable
	}

	if _, err := service.meeting.CancelBooking(ctx, command.BookingID, command.UserID); err != nil {
		return err
	}

	recipient := booking.ClientUserID
	actor := booking.TrainerUserID
	if command.UserID == booking.ClientUserID {
		recipient = booking.TrainerUserID
		actor = booking.ClientUserID
	}
	_ = service.createNotification(ctx, domain.Notification{
		UserID:      recipient,
		Type:        domain.NotificationTypeMeeting,
		ActorUserID: actor,
		Title:       "Занятие отменено",
		Body:        "Запланированное занятие было отменено",
	})

	return nil
}

func (service *Service) ListMeetings(ctx context.Context, query ListMeetingsQuery) ([]domain.MeetingBooking, error) {
	if query.UserID <= 0 {
		return nil, ErrInvalidUserID
	}

	return service.meeting.ListUserBookings(ctx, query.UserID)
}

func (service *Service) isMeetingSlotOffered(ctx context.Context, trainerUserID int64, startsAt time.Time) (bool, error) {
	weekday := int32(startsAt.Weekday())
	hour := int32(startsAt.Hour())

	rules, err := service.meeting.ListMeetingAvailabilityRules(ctx, trainerUserID)
	if err != nil {
		return false, err
	}
	for _, rule := range rules {
		if rule.Weekday == weekday && rule.StartHour == hour {
			return true, nil
		}
	}

	slots, err := service.meeting.ListMeetingSlots(ctx, trainerUserID, startsAt, startsAt.Add(meetingSlotDuration))
	if err != nil {
		return false, err
	}
	for _, slot := range slots {
		if slot.StartsAt.UTC().Equal(startsAt) {
			return true, nil
		}
	}

	return false, nil
}

func normalizeMeetingStart(startsAt time.Time) (time.Time, error) {
	if startsAt.IsZero() {
		return time.Time{}, ErrInvalidMeetingTime
	}
	startsAt = startsAt.UTC()
	if startsAt.Minute() != 0 || startsAt.Second() != 0 || startsAt.Nanosecond() != 0 {
		return time.Time{}, ErrInvalidMeetingTime
	}
	if !startsAt.After(time.Now().UTC()) {
		return time.Time{}, ErrMeetingTimeInPast
	}

	return startsAt, nil
}

func parseMeetingDate(value string, fallback time.Time) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, ErrInvalidMeetingRange
	}

	return parsed.UTC(), nil
}

func overlapsAnyBooking(start time.Time, end time.Time, bookings []domain.MeetingBooking) bool {
	for _, booking := range bookings {
		if booking.StartsAt.Before(end) && booking.EndsAt.After(start) {
			return true
		}
	}

	return false
}
