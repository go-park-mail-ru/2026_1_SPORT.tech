package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) ListMyAvailabilityRules(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.MeetingAvailabilityRulesResponse, error) {
	userID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListMeetingAvailabilityRules(
		forwardContext(ctx),
		&contentv1.ListMeetingAvailabilityRulesRequest{TrainerUserId: userID},
	)
	if err != nil {
		return nil, err
	}

	return mappers.MeetingRulesResponseFromContent(response), nil
}

func (server *Server) CreateAvailabilityRule(ctx context.Context, request *gatewayv1.CreateAvailabilityRuleRequest) (*gatewayv1.MeetingAvailabilityRule, error) {
	userID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.CreateMeetingAvailabilityRule(
		forwardContext(ctx),
		&contentv1.CreateMeetingAvailabilityRuleRequest{
			TrainerUserId: userID,
			Weekday:       request.GetWeekday(),
			StartHour:     request.GetStartHour(),
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.MeetingRuleFromContent(response), nil
}

func (server *Server) DeleteAvailabilityRule(ctx context.Context, request *gatewayv1.DeleteAvailabilityRuleRequest) (*emptypb.Empty, error) {
	userID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := server.contentClient.DeleteMeetingAvailabilityRule(
		forwardContext(ctx),
		&contentv1.DeleteMeetingAvailabilityRuleRequest{
			TrainerUserId: userID,
			RuleId:        request.GetRuleId(),
		},
	); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) CreateAvailabilitySlot(ctx context.Context, request *gatewayv1.CreateAvailabilitySlotRequest) (*gatewayv1.MeetingSlot, error) {
	userID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.CreateMeetingSlot(
		forwardContext(ctx),
		&contentv1.CreateMeetingSlotRequest{
			TrainerUserId: userID,
			StartsAt:      request.GetStartsAt(),
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.MeetingSlotFromContent(response), nil
}

func (server *Server) ListMyAvailabilitySlots(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.MeetingSlotsResponse, error) {
	userID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListMeetingSlots(
		forwardContext(ctx),
		&contentv1.ListMeetingSlotsRequest{TrainerUserId: userID},
	)
	if err != nil {
		return nil, err
	}

	return mappers.MeetingSlotsResponseFromContent(response), nil
}

func (server *Server) DeleteAvailabilitySlot(ctx context.Context, request *gatewayv1.DeleteAvailabilitySlotRequest) (*emptypb.Empty, error) {
	userID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := server.contentClient.DeleteMeetingSlot(
		forwardContext(ctx),
		&contentv1.DeleteMeetingSlotRequest{
			TrainerUserId: userID,
			SlotId:        request.GetSlotId(),
		},
	); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) GetTrainerAvailability(ctx context.Context, request *gatewayv1.GetTrainerAvailabilityRequest) (*gatewayv1.MeetingAvailabilityResponse, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListTrainerMeetingAvailability(
		forwardContext(ctx),
		&contentv1.ListTrainerMeetingAvailabilityRequest{
			TrainerUserId: request.GetTrainerId(),
			ViewerUserId:  userID,
			From:          request.GetFrom(),
			To:            request.GetTo(),
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.MeetingAvailabilityResponseFromContent(response), nil
}

func (server *Server) BookMeeting(ctx context.Context, request *gatewayv1.BookMeetingRequest) (*gatewayv1.MeetingBooking, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.BookMeeting(
		forwardContext(ctx),
		&contentv1.BookMeetingRequest{
			ClientUserId:  userID,
			TrainerUserId: request.GetTrainerId(),
			StartsAt:      request.GetStartsAt(),
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.MeetingBookingFromContent(response, userID), nil
}

func (server *Server) AssignMeeting(ctx context.Context, request *gatewayv1.AssignMeetingRequest) (*gatewayv1.MeetingBooking, error) {
	userID, err := server.requireTrainerUserID(ctx)
	if err != nil {
		return nil, err
	}

	contentRequest := &contentv1.AssignMeetingRequest{
		TrainerUserId: userID,
		ClientUserId:  request.GetClientUserId(),
		StartsAt:      request.GetStartsAt(),
		DurationHours: request.GetDurationHours(),
	}
	if request.Note != nil {
		contentRequest.Note = request.Note
	}

	response, err := server.contentClient.AssignMeeting(forwardContext(ctx), contentRequest)
	if err != nil {
		return nil, err
	}

	return mappers.MeetingBookingFromContent(response, userID), nil
}

func (server *Server) ListMyMeetings(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.MeetingsResponse, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListMeetings(
		forwardContext(ctx),
		&contentv1.ListMeetingsRequest{UserId: userID},
	)
	if err != nil {
		return nil, err
	}

	return mappers.MeetingsResponseFromContent(response, userID), nil
}

func (server *Server) CancelMeeting(ctx context.Context, request *gatewayv1.CancelMeetingRequest) (*emptypb.Empty, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := server.contentClient.CancelMeeting(
		forwardContext(ctx),
		&contentv1.CancelMeetingRequest{
			UserId:    userID,
			BookingId: request.GetBookingId(),
		},
	); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
