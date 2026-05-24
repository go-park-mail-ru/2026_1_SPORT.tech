package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func MeetingRuleFromContent(rule *contentv1.MeetingAvailabilityRule) *gatewayv1.MeetingAvailabilityRule {
	if rule == nil {
		return nil
	}

	return &gatewayv1.MeetingAvailabilityRule{
		RuleId:    rule.GetRuleId(),
		Weekday:   rule.GetWeekday(),
		StartHour: rule.GetStartHour(),
		CreatedAt: rule.GetCreatedAt(),
	}
}

func MeetingRulesResponseFromContent(response *contentv1.ListMeetingAvailabilityRulesResponse) *gatewayv1.MeetingAvailabilityRulesResponse {
	rules := make([]*gatewayv1.MeetingAvailabilityRule, 0)
	if response != nil {
		rules = make([]*gatewayv1.MeetingAvailabilityRule, 0, len(response.GetRules()))
		for _, rule := range response.GetRules() {
			rules = append(rules, MeetingRuleFromContent(rule))
		}
	}

	return &gatewayv1.MeetingAvailabilityRulesResponse{Rules: rules}
}

func MeetingSlotFromContent(slot *contentv1.MeetingSlot) *gatewayv1.MeetingSlot {
	if slot == nil {
		return nil
	}

	return &gatewayv1.MeetingSlot{
		SlotId:    slot.GetSlotId(),
		StartsAt:  slot.GetStartsAt(),
		CreatedAt: slot.GetCreatedAt(),
	}
}

func MeetingAvailabilityResponseFromContent(response *contentv1.ListTrainerMeetingAvailabilityResponse) *gatewayv1.MeetingAvailabilityResponse {
	slots := make([]*gatewayv1.MeetingAvailabilitySlot, 0)
	if response != nil {
		slots = make([]*gatewayv1.MeetingAvailabilitySlot, 0, len(response.GetSlots()))
		for _, slot := range response.GetSlots() {
			slots = append(slots, &gatewayv1.MeetingAvailabilitySlot{
				StartsAt: slot.GetStartsAt(),
				EndsAt:   slot.GetEndsAt(),
			})
		}
	}

	return &gatewayv1.MeetingAvailabilityResponse{Slots: slots}
}

func MeetingBookingFromContent(booking *contentv1.MeetingBooking, requesterUserID int64) *gatewayv1.MeetingBooking {
	if booking == nil {
		return nil
	}

	role := "client"
	otherUserID := booking.GetTrainerUserId()
	if booking.GetTrainerUserId() == requesterUserID {
		role = "trainer"
		otherUserID = booking.GetClientUserId()
	}

	result := &gatewayv1.MeetingBooking{
		BookingId:       booking.GetBookingId(),
		TrainerUserId:   booking.GetTrainerUserId(),
		ClientUserId:    booking.GetClientUserId(),
		StartsAt:        booking.GetStartsAt(),
		EndsAt:          booking.GetEndsAt(),
		Status:          booking.GetStatus(),
		CreatedByUserId: booking.GetCreatedByUserId(),
		CreatedAt:       booking.GetCreatedAt(),
		Role:            role,
		OtherUserId:     otherUserID,
	}
	if booking.Note != nil {
		result.Note = booking.Note
	}

	return result
}

func MeetingsResponseFromContent(response *contentv1.ListMeetingsResponse, requesterUserID int64) *gatewayv1.MeetingsResponse {
	bookings := make([]*gatewayv1.MeetingBooking, 0)
	if response != nil {
		bookings = make([]*gatewayv1.MeetingBooking, 0, len(response.GetBookings()))
		for _, booking := range response.GetBookings() {
			bookings = append(bookings, MeetingBookingFromContent(booking, requesterUserID))
		}
	}

	return &gatewayv1.MeetingsResponse{Bookings: bookings}
}
