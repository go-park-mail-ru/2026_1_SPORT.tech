package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) CreateMeetingAvailabilityRule(ctx context.Context, request *contentv1.CreateMeetingAvailabilityRuleRequest) (*contentv1.MeetingAvailabilityRule, error) {
	rule, err := server.useCases.Meeting.CreateMeetingAvailabilityRule(ctx, usecase.CreateMeetingAvailabilityRuleCommand{
		TrainerUserID: request.GetTrainerUserId(),
		Weekday:       request.GetWeekday(),
		StartHour:     request.GetStartHour(),
	})
	if err != nil {
		return nil, server.statusError("CreateMeetingAvailabilityRule", err)
	}

	return mappers.MeetingRuleToProto(rule), nil
}

func (server *Server) ListMeetingAvailabilityRules(ctx context.Context, request *contentv1.ListMeetingAvailabilityRulesRequest) (*contentv1.ListMeetingAvailabilityRulesResponse, error) {
	rules, err := server.useCases.Meeting.ListMeetingAvailabilityRules(ctx, usecase.ListMeetingAvailabilityRulesQuery{
		TrainerUserID: request.GetTrainerUserId(),
	})
	if err != nil {
		return nil, server.statusError("ListMeetingAvailabilityRules", err)
	}

	return mappers.NewListMeetingAvailabilityRulesResponse(rules), nil
}

func (server *Server) DeleteMeetingAvailabilityRule(ctx context.Context, request *contentv1.DeleteMeetingAvailabilityRuleRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Meeting.DeleteMeetingAvailabilityRule(ctx, usecase.DeleteMeetingAvailabilityRuleCommand{
		TrainerUserID: request.GetTrainerUserId(),
		RuleID:        request.GetRuleId(),
	}); err != nil {
		return nil, server.statusError("DeleteMeetingAvailabilityRule", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) CreateMeetingSlot(ctx context.Context, request *contentv1.CreateMeetingSlotRequest) (*contentv1.MeetingSlot, error) {
	slot, err := server.useCases.Meeting.CreateMeetingSlot(ctx, usecase.CreateMeetingSlotCommand{
		TrainerUserID: request.GetTrainerUserId(),
		StartsAt:      request.GetStartsAt().AsTime(),
	})
	if err != nil {
		return nil, server.statusError("CreateMeetingSlot", err)
	}

	return mappers.MeetingSlotToProto(slot), nil
}

func (server *Server) DeleteMeetingSlot(ctx context.Context, request *contentv1.DeleteMeetingSlotRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Meeting.DeleteMeetingSlot(ctx, usecase.DeleteMeetingSlotCommand{
		TrainerUserID: request.GetTrainerUserId(),
		SlotID:        request.GetSlotId(),
	}); err != nil {
		return nil, server.statusError("DeleteMeetingSlot", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) ListTrainerMeetingAvailability(ctx context.Context, request *contentv1.ListTrainerMeetingAvailabilityRequest) (*contentv1.ListTrainerMeetingAvailabilityResponse, error) {
	slots, err := server.useCases.Meeting.ListTrainerMeetingAvailability(ctx, usecase.ListTrainerMeetingAvailabilityQuery{
		TrainerUserID: request.GetTrainerUserId(),
		ViewerUserID:  request.GetViewerUserId(),
		From:          request.GetFrom(),
		To:            request.GetTo(),
	})
	if err != nil {
		return nil, server.statusError("ListTrainerMeetingAvailability", err)
	}

	return mappers.NewListTrainerMeetingAvailabilityResponse(slots), nil
}

func (server *Server) BookMeeting(ctx context.Context, request *contentv1.BookMeetingRequest) (*contentv1.MeetingBooking, error) {
	booking, err := server.useCases.Meeting.BookMeeting(ctx, usecase.BookMeetingCommand{
		ClientUserID:  request.GetClientUserId(),
		TrainerUserID: request.GetTrainerUserId(),
		StartsAt:      request.GetStartsAt().AsTime(),
	})
	if err != nil {
		return nil, server.statusError("BookMeeting", err)
	}

	return mappers.MeetingBookingToProto(booking), nil
}

func (server *Server) AssignMeeting(ctx context.Context, request *contentv1.AssignMeetingRequest) (*contentv1.MeetingBooking, error) {
	command := usecase.AssignMeetingCommand{
		TrainerUserID: request.GetTrainerUserId(),
		ClientUserID:  request.GetClientUserId(),
		StartsAt:      request.GetStartsAt().AsTime(),
		DurationHours: request.GetDurationHours(),
	}
	if request.Note != nil {
		command.Note = request.Note
	}

	booking, err := server.useCases.Meeting.AssignMeeting(ctx, command)
	if err != nil {
		return nil, server.statusError("AssignMeeting", err)
	}

	return mappers.MeetingBookingToProto(booking), nil
}

func (server *Server) CancelMeeting(ctx context.Context, request *contentv1.CancelMeetingRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Meeting.CancelMeeting(ctx, usecase.CancelMeetingCommand{
		UserID:    request.GetUserId(),
		BookingID: request.GetBookingId(),
	}); err != nil {
		return nil, server.statusError("CancelMeeting", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) ListMeetings(ctx context.Context, request *contentv1.ListMeetingsRequest) (*contentv1.ListMeetingsResponse, error) {
	bookings, err := server.useCases.Meeting.ListMeetings(ctx, usecase.ListMeetingsQuery{
		UserID: request.GetUserId(),
	})
	if err != nil {
		return nil, server.statusError("ListMeetings", err)
	}

	return mappers.NewListMeetingsResponse(bookings), nil
}
