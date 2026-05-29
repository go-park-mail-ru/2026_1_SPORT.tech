package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

type stubMeetingRepository struct {
	hasCalendarSubFunc func(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error)
	createRuleFunc     func(ctx context.Context, rule domain.MeetingAvailabilityRule) (domain.MeetingAvailabilityRule, error)
	listRulesFunc      func(ctx context.Context, trainerUserID int64) ([]domain.MeetingAvailabilityRule, error)
	deleteRuleFunc     func(ctx context.Context, trainerUserID int64, ruleID int64) error
	createSlotFunc     func(ctx context.Context, slot domain.MeetingSlot) (domain.MeetingSlot, error)
	deleteSlotFunc     func(ctx context.Context, trainerUserID int64, slotID int64) error
	listSlotsFunc      func(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingSlot, error)
	listBookingsFunc   func(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingBooking, error)
	countActiveFunc    func(ctx context.Context, clientUserID int64, trainerUserID int64, now time.Time) (int32, error)
	createBookingFunc  func(ctx context.Context, booking domain.MeetingBooking) (domain.MeetingBooking, error)
	getBookingFunc     func(ctx context.Context, bookingID int64) (domain.MeetingBooking, error)
	cancelBookingFunc  func(ctx context.Context, bookingID int64, cancelledByUserID int64) (domain.MeetingBooking, error)
	listUserFunc       func(ctx context.Context, userID int64) ([]domain.MeetingBooking, error)
}

func (repository stubMeetingRepository) HasActiveCalendarSubscription(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
	if repository.hasCalendarSubFunc == nil {
		return false, nil
	}
	return repository.hasCalendarSubFunc(ctx, clientUserID, trainerUserID)
}

func (repository stubMeetingRepository) CreateMeetingAvailabilityRule(ctx context.Context, rule domain.MeetingAvailabilityRule) (domain.MeetingAvailabilityRule, error) {
	if repository.createRuleFunc == nil {
		return rule, nil
	}
	return repository.createRuleFunc(ctx, rule)
}

func (repository stubMeetingRepository) ListMeetingAvailabilityRules(ctx context.Context, trainerUserID int64) ([]domain.MeetingAvailabilityRule, error) {
	if repository.listRulesFunc == nil {
		return nil, nil
	}
	return repository.listRulesFunc(ctx, trainerUserID)
}

func (repository stubMeetingRepository) DeleteMeetingAvailabilityRule(ctx context.Context, trainerUserID int64, ruleID int64) error {
	if repository.deleteRuleFunc == nil {
		return nil
	}
	return repository.deleteRuleFunc(ctx, trainerUserID, ruleID)
}

func (repository stubMeetingRepository) CreateMeetingSlot(ctx context.Context, slot domain.MeetingSlot) (domain.MeetingSlot, error) {
	if repository.createSlotFunc == nil {
		return slot, nil
	}
	return repository.createSlotFunc(ctx, slot)
}

func (repository stubMeetingRepository) DeleteMeetingSlot(ctx context.Context, trainerUserID int64, slotID int64) error {
	if repository.deleteSlotFunc == nil {
		return nil
	}
	return repository.deleteSlotFunc(ctx, trainerUserID, slotID)
}

func (repository stubMeetingRepository) ListMeetingSlots(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingSlot, error) {
	if repository.listSlotsFunc == nil {
		return nil, nil
	}
	return repository.listSlotsFunc(ctx, trainerUserID, from, to)
}

func (repository stubMeetingRepository) ListConfirmedBookings(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingBooking, error) {
	if repository.listBookingsFunc == nil {
		return nil, nil
	}
	return repository.listBookingsFunc(ctx, trainerUserID, from, to)
}

func (repository stubMeetingRepository) CountActiveClientBookings(ctx context.Context, clientUserID int64, trainerUserID int64, now time.Time) (int32, error) {
	if repository.countActiveFunc == nil {
		return 0, nil
	}
	return repository.countActiveFunc(ctx, clientUserID, trainerUserID, now)
}

func (repository stubMeetingRepository) CreateBooking(ctx context.Context, booking domain.MeetingBooking) (domain.MeetingBooking, error) {
	if repository.createBookingFunc == nil {
		booking.BookingID = 1
		return booking, nil
	}
	return repository.createBookingFunc(ctx, booking)
}

func (repository stubMeetingRepository) GetBooking(ctx context.Context, bookingID int64) (domain.MeetingBooking, error) {
	if repository.getBookingFunc == nil {
		return domain.MeetingBooking{}, domain.ErrMeetingBookingNotFound
	}
	return repository.getBookingFunc(ctx, bookingID)
}

func (repository stubMeetingRepository) CancelBooking(ctx context.Context, bookingID int64, cancelledByUserID int64) (domain.MeetingBooking, error) {
	if repository.cancelBookingFunc == nil {
		return domain.MeetingBooking{}, nil
	}
	return repository.cancelBookingFunc(ctx, bookingID, cancelledByUserID)
}

func (repository stubMeetingRepository) ListUserBookings(ctx context.Context, userID int64) ([]domain.MeetingBooking, error) {
	if repository.listUserFunc == nil {
		return nil, nil
	}
	return repository.listUserFunc(ctx, userID)
}

func newMeetingService(repository stubMeetingRepository) *Service {
	return NewService(Repositories{Meeting: repository}, nil)
}

func futureHour(hoursFromNow int) time.Time {
	return time.Now().UTC().Truncate(time.Hour).Add(time.Duration(hoursFromNow) * time.Hour)
}

func TestCreateMeetingAvailabilityRule(t *testing.T) {
	service := newMeetingService(stubMeetingRepository{
		createRuleFunc: func(ctx context.Context, rule domain.MeetingAvailabilityRule) (domain.MeetingAvailabilityRule, error) {
			rule.RuleID = 7
			return rule, nil
		},
	})

	rule, err := service.CreateMeetingAvailabilityRule(context.Background(), CreateMeetingAvailabilityRuleCommand{
		TrainerUserID: 10,
		Weekday:       3,
		StartHour:     9,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.RuleID != 7 {
		t.Fatalf("expected rule id 7, got %d", rule.RuleID)
	}
}

func TestCreateMeetingAvailabilityRuleValidation(t *testing.T) {
	service := newMeetingService(stubMeetingRepository{})

	cases := []struct {
		name    string
		command CreateMeetingAvailabilityRuleCommand
		wantErr error
	}{
		{"invalid user", CreateMeetingAvailabilityRuleCommand{TrainerUserID: 0, Weekday: 1, StartHour: 9}, ErrInvalidUserID},
		{"invalid weekday", CreateMeetingAvailabilityRuleCommand{TrainerUserID: 1, Weekday: 7, StartHour: 9}, ErrInvalidMeetingWeekday},
		{"invalid hour", CreateMeetingAvailabilityRuleCommand{TrainerUserID: 1, Weekday: 1, StartHour: 24}, ErrInvalidMeetingHour},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := service.CreateMeetingAvailabilityRule(context.Background(), testCase.command); !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected %v, got %v", testCase.wantErr, err)
			}
		})
	}
}

func TestListAndDeleteMeetingAvailabilityRule(t *testing.T) {
	service := newMeetingService(stubMeetingRepository{
		listRulesFunc: func(ctx context.Context, trainerUserID int64) ([]domain.MeetingAvailabilityRule, error) {
			return []domain.MeetingAvailabilityRule{{RuleID: 1, TrainerUserID: trainerUserID}}, nil
		},
	})

	rules, err := service.ListMeetingAvailabilityRules(context.Background(), ListMeetingAvailabilityRulesQuery{TrainerUserID: 5})
	if err != nil || len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d (err %v)", len(rules), err)
	}

	if _, err := service.ListMeetingAvailabilityRules(context.Background(), ListMeetingAvailabilityRulesQuery{TrainerUserID: 0}); !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}

	if err := service.DeleteMeetingAvailabilityRule(context.Background(), DeleteMeetingAvailabilityRuleCommand{TrainerUserID: 5, RuleID: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := service.DeleteMeetingAvailabilityRule(context.Background(), DeleteMeetingAvailabilityRuleCommand{TrainerUserID: 5, RuleID: 0}); !errors.Is(err, ErrInvalidMeetingRuleID) {
		t.Fatalf("expected ErrInvalidMeetingRuleID, got %v", err)
	}
}

func TestCreateMeetingSlotValidation(t *testing.T) {
	service := newMeetingService(stubMeetingRepository{})

	if _, err := service.CreateMeetingSlot(context.Background(), CreateMeetingSlotCommand{TrainerUserID: 1, StartsAt: futureHour(2).Add(30 * time.Minute)}); !errors.Is(err, ErrInvalidMeetingTime) {
		t.Fatalf("expected ErrInvalidMeetingTime, got %v", err)
	}
	if _, err := service.CreateMeetingSlot(context.Background(), CreateMeetingSlotCommand{TrainerUserID: 1, StartsAt: futureHour(-5)}); !errors.Is(err, ErrMeetingTimeInPast) {
		t.Fatalf("expected ErrMeetingTimeInPast, got %v", err)
	}
	slot, err := service.CreateMeetingSlot(context.Background(), CreateMeetingSlotCommand{TrainerUserID: 1, StartsAt: futureHour(3)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slot.StartsAt.Equal(futureHour(3)) {
		t.Fatalf("unexpected slot start %v", slot.StartsAt)
	}
}

func TestListMyMeetingSlots(t *testing.T) {
	slotStart := futureHour(48)
	service := newMeetingService(stubMeetingRepository{
		listSlotsFunc: func(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingSlot, error) {
			return []domain.MeetingSlot{{SlotID: 11, TrainerUserID: trainerUserID, StartsAt: slotStart}}, nil
		},
	})

	slots, err := service.ListMyMeetingSlots(context.Background(), ListMyMeetingSlotsQuery{TrainerUserID: 4})
	if err != nil || len(slots) != 1 || slots[0].SlotID != 11 {
		t.Fatalf("expected 1 slot id 11, got %+v (err %v)", slots, err)
	}

	if _, err := service.ListMyMeetingSlots(context.Background(), ListMyMeetingSlotsQuery{TrainerUserID: 0}); !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestListTrainerMeetingAvailabilityAccess(t *testing.T) {
	service := newMeetingService(stubMeetingRepository{
		hasCalendarSubFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
			return false, nil
		},
	})

	if _, err := service.ListTrainerMeetingAvailability(context.Background(), ListTrainerMeetingAvailabilityQuery{
		TrainerUserID: 1,
		ViewerUserID:  2,
	}); !errors.Is(err, domain.ErrMeetingAccessForbidden) {
		t.Fatalf("expected ErrMeetingAccessForbidden, got %v", err)
	}
}

func TestListTrainerMeetingAvailabilityGeneratesSlots(t *testing.T) {
	openSlot := futureHour(48)
	bookedSlot := futureHour(72)

	service := newMeetingService(stubMeetingRepository{
		listSlotsFunc: func(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingSlot, error) {
			return []domain.MeetingSlot{
				{SlotID: 1, TrainerUserID: trainerUserID, StartsAt: openSlot},
				{SlotID: 2, TrainerUserID: trainerUserID, StartsAt: bookedSlot},
			}, nil
		},
		listBookingsFunc: func(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingBooking, error) {
			return []domain.MeetingBooking{
				{StartsAt: bookedSlot, EndsAt: bookedSlot.Add(time.Hour), Status: domain.MeetingBookingStatusConfirmed},
			}, nil
		},
	})

	slots, err := service.ListTrainerMeetingAvailability(context.Background(), ListTrainerMeetingAvailabilityQuery{
		TrainerUserID: 1,
		ViewerUserID:  1,
		From:          time.Now().UTC().Format("2006-01-02"),
		To:            time.Now().UTC().AddDate(0, 0, 10).Format("2006-01-02"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 1 {
		t.Fatalf("expected 1 open slot (booked excluded), got %d", len(slots))
	}
	if !slots[0].StartsAt.Equal(openSlot) {
		t.Fatalf("expected open slot %v, got %v", openSlot, slots[0].StartsAt)
	}
}

func TestListTrainerMeetingAvailabilityInvalidRange(t *testing.T) {
	service := newMeetingService(stubMeetingRepository{})
	_, err := service.ListTrainerMeetingAvailability(context.Background(), ListTrainerMeetingAvailabilityQuery{
		TrainerUserID: 1,
		ViewerUserID:  1,
		From:          "2026-05-10",
		To:            "2026-05-01",
	})
	if !errors.Is(err, ErrInvalidMeetingRange) {
		t.Fatalf("expected ErrInvalidMeetingRange, got %v", err)
	}
}

func TestBookMeetingSuccess(t *testing.T) {
	start := futureHour(24)
	notifications := make([]domain.Notification, 0, 2)
	service := NewService(Repositories{
		Meeting: stubMeetingRepository{
			hasCalendarSubFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
				return true, nil
			},
			listSlotsFunc: func(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingSlot, error) {
				return []domain.MeetingSlot{{StartsAt: start}}, nil
			},
			createBookingFunc: func(ctx context.Context, booking domain.MeetingBooking) (domain.MeetingBooking, error) {
				booking.BookingID = 42
				return booking, nil
			},
		},
		Notifications: notifyStub{createFunc: func(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
			notifications = append(notifications, notification)
			return notification, nil
		}},
	}, nil)

	booking, err := service.BookMeeting(context.Background(), BookMeetingCommand{
		ClientUserID:  2,
		TrainerUserID: 1,
		StartsAt:      start,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if booking.BookingID != 42 {
		t.Fatalf("expected booking id 42, got %d", booking.BookingID)
	}
	if !booking.EndsAt.Equal(start.Add(time.Hour)) {
		t.Fatalf("expected 1h booking, got %v-%v", booking.StartsAt, booking.EndsAt)
	}
	if len(notifications) != 2 {
		t.Fatalf("expected trainer and client notifications, got %+v", notifications)
	}
	if notifications[0].UserID != 1 ||
		notifications[0].ActorUserID != 2 ||
		notifications[0].Type != domain.NotificationTypeMeeting {
		t.Fatalf("unexpected trainer notification: %+v", notifications[0])
	}
	if notifications[1].UserID != 2 ||
		notifications[1].ActorUserID != 1 ||
		notifications[1].Type != domain.NotificationTypeMeeting ||
		notifications[1].Title != "Запись подтверждена" {
		t.Fatalf("unexpected client notification: %+v", notifications[1])
	}
}

func TestBookMeetingForbiddenWithoutSubscription(t *testing.T) {
	start := futureHour(24)
	service := newMeetingService(stubMeetingRepository{
		hasCalendarSubFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
			return false, nil
		},
	})

	if _, err := service.BookMeeting(context.Background(), BookMeetingCommand{ClientUserID: 2, TrainerUserID: 1, StartsAt: start}); !errors.Is(err, domain.ErrMeetingAccessForbidden) {
		t.Fatalf("expected ErrMeetingAccessForbidden, got %v", err)
	}
}

func TestBookMeetingSlotNotOffered(t *testing.T) {
	start := futureHour(24)
	service := newMeetingService(stubMeetingRepository{
		hasCalendarSubFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
			return true, nil
		},
	})

	if _, err := service.BookMeeting(context.Background(), BookMeetingCommand{ClientUserID: 2, TrainerUserID: 1, StartsAt: start}); !errors.Is(err, domain.ErrMeetingSlotUnavailable) {
		t.Fatalf("expected ErrMeetingSlotUnavailable, got %v", err)
	}
}

func TestBookMeetingClientLimit(t *testing.T) {
	start := futureHour(24)
	service := newMeetingService(stubMeetingRepository{
		hasCalendarSubFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
			return true, nil
		},
		listSlotsFunc: func(ctx context.Context, trainerUserID int64, from time.Time, to time.Time) ([]domain.MeetingSlot, error) {
			return []domain.MeetingSlot{{StartsAt: start}}, nil
		},
		countActiveFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64, now time.Time) (int32, error) {
			return 1, nil
		},
	})

	if _, err := service.BookMeeting(context.Background(), BookMeetingCommand{ClientUserID: 2, TrainerUserID: 1, StartsAt: start}); !errors.Is(err, domain.ErrMeetingClientLimit) {
		t.Fatalf("expected ErrMeetingClientLimit, got %v", err)
	}
}

func TestAssignMeetingSuccess(t *testing.T) {
	start := futureHour(24)
	note := "bring resistance bands"
	service := newMeetingService(stubMeetingRepository{
		hasCalendarSubFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
			return true, nil
		},
		createBookingFunc: func(ctx context.Context, booking domain.MeetingBooking) (domain.MeetingBooking, error) {
			booking.BookingID = 9
			return booking, nil
		},
	})

	booking, err := service.AssignMeeting(context.Background(), AssignMeetingCommand{
		TrainerUserID: 1,
		ClientUserID:  2,
		StartsAt:      start,
		DurationHours: 2,
		Note:          &note,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !booking.EndsAt.Equal(start.Add(2 * time.Hour)) {
		t.Fatalf("expected 2h booking, got %v-%v", booking.StartsAt, booking.EndsAt)
	}
	if booking.CreatedByUserID != 1 {
		t.Fatalf("expected created by trainer, got %d", booking.CreatedByUserID)
	}
}

func TestAssignMeetingInvalidDuration(t *testing.T) {
	start := futureHour(24)
	service := newMeetingService(stubMeetingRepository{})
	if _, err := service.AssignMeeting(context.Background(), AssignMeetingCommand{TrainerUserID: 1, ClientUserID: 2, StartsAt: start, DurationHours: 0}); !errors.Is(err, ErrInvalidMeetingDuration) {
		t.Fatalf("expected ErrInvalidMeetingDuration, got %v", err)
	}
	if _, err := service.AssignMeeting(context.Background(), AssignMeetingCommand{TrainerUserID: 1, ClientUserID: 2, StartsAt: start, DurationHours: 99}); !errors.Is(err, ErrInvalidMeetingDuration) {
		t.Fatalf("expected ErrInvalidMeetingDuration, got %v", err)
	}
}

func TestAssignMeetingForbiddenWithoutSubscription(t *testing.T) {
	start := futureHour(24)
	service := newMeetingService(stubMeetingRepository{
		hasCalendarSubFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (bool, error) {
			return false, nil
		},
	})
	if _, err := service.AssignMeeting(context.Background(), AssignMeetingCommand{TrainerUserID: 1, ClientUserID: 2, StartsAt: start, DurationHours: 1}); !errors.Is(err, domain.ErrMeetingAccessForbidden) {
		t.Fatalf("expected ErrMeetingAccessForbidden, got %v", err)
	}
}

func TestCancelMeeting(t *testing.T) {
	start := futureHour(24)
	booking := domain.MeetingBooking{
		BookingID:     5,
		TrainerUserID: 1,
		ClientUserID:  2,
		StartsAt:      start,
		EndsAt:        start.Add(time.Hour),
		Status:        domain.MeetingBookingStatusConfirmed,
	}

	service := newMeetingService(stubMeetingRepository{
		getBookingFunc: func(ctx context.Context, bookingID int64) (domain.MeetingBooking, error) {
			return booking, nil
		},
		cancelBookingFunc: func(ctx context.Context, bookingID int64, cancelledByUserID int64) (domain.MeetingBooking, error) {
			booking.Status = domain.MeetingBookingStatusCancelled
			return booking, nil
		},
	})

	if err := service.CancelMeeting(context.Background(), CancelMeetingCommand{UserID: 2, BookingID: 5}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCancelMeetingNotParticipant(t *testing.T) {
	start := futureHour(24)
	service := newMeetingService(stubMeetingRepository{
		getBookingFunc: func(ctx context.Context, bookingID int64) (domain.MeetingBooking, error) {
			return domain.MeetingBooking{BookingID: 5, TrainerUserID: 1, ClientUserID: 2, StartsAt: start, Status: domain.MeetingBookingStatusConfirmed}, nil
		},
	})
	if err := service.CancelMeeting(context.Background(), CancelMeetingCommand{UserID: 99, BookingID: 5}); !errors.Is(err, domain.ErrMeetingBookingNotFound) {
		t.Fatalf("expected ErrMeetingBookingNotFound, got %v", err)
	}
}

func TestCancelMeetingPastNotCancelable(t *testing.T) {
	service := newMeetingService(stubMeetingRepository{
		getBookingFunc: func(ctx context.Context, bookingID int64) (domain.MeetingBooking, error) {
			return domain.MeetingBooking{BookingID: 5, TrainerUserID: 1, ClientUserID: 2, StartsAt: futureHour(-10), Status: domain.MeetingBookingStatusConfirmed}, nil
		},
	})
	if err := service.CancelMeeting(context.Background(), CancelMeetingCommand{UserID: 1, BookingID: 5}); !errors.Is(err, domain.ErrMeetingNotCancelable) {
		t.Fatalf("expected ErrMeetingNotCancelable, got %v", err)
	}
}

func TestListMeetings(t *testing.T) {
	service := newMeetingService(stubMeetingRepository{
		listUserFunc: func(ctx context.Context, userID int64) ([]domain.MeetingBooking, error) {
			return []domain.MeetingBooking{{BookingID: 1}, {BookingID: 2}}, nil
		},
	})

	bookings, err := service.ListMeetings(context.Background(), ListMeetingsQuery{UserID: 3})
	if err != nil || len(bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d (err %v)", len(bookings), err)
	}
	if _, err := service.ListMeetings(context.Background(), ListMeetingsQuery{UserID: 0}); !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}
}

type notifyStub struct {
	createFunc func(ctx context.Context, notification domain.Notification) (domain.Notification, error)
}

func (stub notifyStub) CreateNotification(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
	return stub.createFunc(ctx, notification)
}

func (stub notifyStub) ListNotifications(ctx context.Context, userID int64, limit int32, offset int32) ([]domain.Notification, error) {
	return nil, nil
}

func (stub notifyStub) MarkNotificationRead(ctx context.Context, userID int64, notificationID int64) (domain.Notification, error) {
	return domain.Notification{}, nil
}
