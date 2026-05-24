package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MeetingRuleToProto(rule domain.MeetingAvailabilityRule) *contentv1.MeetingAvailabilityRule {
	return &contentv1.MeetingAvailabilityRule{
		RuleId:        rule.RuleID,
		TrainerUserId: rule.TrainerUserID,
		Weekday:       rule.Weekday,
		StartHour:     rule.StartHour,
		CreatedAt:     timestamppb.New(rule.CreatedAt),
	}
}

func NewListMeetingAvailabilityRulesResponse(rules []domain.MeetingAvailabilityRule) *contentv1.ListMeetingAvailabilityRulesResponse {
	response := &contentv1.ListMeetingAvailabilityRulesResponse{
		Rules: make([]*contentv1.MeetingAvailabilityRule, 0, len(rules)),
	}
	for _, rule := range rules {
		response.Rules = append(response.Rules, MeetingRuleToProto(rule))
	}

	return response
}

func MeetingSlotToProto(slot domain.MeetingSlot) *contentv1.MeetingSlot {
	return &contentv1.MeetingSlot{
		SlotId:        slot.SlotID,
		TrainerUserId: slot.TrainerUserID,
		StartsAt:      timestamppb.New(slot.StartsAt),
		CreatedAt:     timestamppb.New(slot.CreatedAt),
	}
}

func NewListTrainerMeetingAvailabilityResponse(slots []domain.MeetingAvailabilitySlot) *contentv1.ListTrainerMeetingAvailabilityResponse {
	response := &contentv1.ListTrainerMeetingAvailabilityResponse{
		Slots: make([]*contentv1.MeetingAvailabilitySlot, 0, len(slots)),
	}
	for _, slot := range slots {
		response.Slots = append(response.Slots, &contentv1.MeetingAvailabilitySlot{
			StartsAt: timestamppb.New(slot.StartsAt),
			EndsAt:   timestamppb.New(slot.EndsAt),
		})
	}

	return response
}

func MeetingBookingToProto(booking domain.MeetingBooking) *contentv1.MeetingBooking {
	response := &contentv1.MeetingBooking{
		BookingId:       booking.BookingID,
		TrainerUserId:   booking.TrainerUserID,
		ClientUserId:    booking.ClientUserID,
		StartsAt:        timestamppb.New(booking.StartsAt),
		EndsAt:          timestamppb.New(booking.EndsAt),
		Status:          string(booking.Status),
		CreatedByUserId: booking.CreatedByUserID,
		CreatedAt:       timestamppb.New(booking.CreatedAt),
	}
	if booking.Note != nil {
		response.Note = booking.Note
	}

	return response
}

func NewListMeetingsResponse(bookings []domain.MeetingBooking) *contentv1.ListMeetingsResponse {
	response := &contentv1.ListMeetingsResponse{
		Bookings: make([]*contentv1.MeetingBooking, 0, len(bookings)),
	}
	for _, booking := range bookings {
		response.Bookings = append(response.Bookings, MeetingBookingToProto(booking))
	}

	return response
}
