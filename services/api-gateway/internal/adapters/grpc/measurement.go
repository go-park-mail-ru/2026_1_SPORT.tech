package grpc

import (
	"context"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) CreateMyMeasurement(ctx context.Context, request *gatewayv1.CreateMeasurementRequest) (*gatewayv1.MeasurementResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	resp, err := server.profileClient.CreateMeasurement(forwardContext(ctx), &profilev1.CreateMeasurementRequest{
		UserId:     userID,
		MeasuredAt: request.GetMeasuredAt(),
		WeightKg:   request.WeightKg,
		BodyFatPct: request.BodyFatPct,
		ChestCm:    request.ChestCm,
		WaistCm:    request.WaistCm,
		HipsCm:     request.HipsCm,
		Notes:      request.Notes,
	})
	if err != nil {
		return nil, err
	}

	return measurementResponseFromProfile(resp), nil
}

func (server *Server) ListMeasurements(ctx context.Context, request *gatewayv1.ListMeasurementsRequest) (*gatewayv1.ListMeasurementsResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	if principal == nil || principal.User == nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	viewerID := principal.User.GetUserId()
	userID := request.GetUserId()
	if userID == 0 {
		userID = viewerID
	}

	// Если смотрим свой профиль, viewer = 0 (без ограничений).
	passViewerID := viewerID
	if viewerID == userID {
		passViewerID = 0
	}

	resp, err := server.profileClient.ListMeasurements(forwardContext(ctx), &profilev1.ListMeasurementsRequest{
		UserId:       userID,
		Limit:        request.GetLimit(),
		Offset:       request.GetOffset(),
		ViewerUserId: passViewerID,
	})
	if err != nil {
		return nil, err
	}

	out := &gatewayv1.ListMeasurementsResponse{
		Measurements: make([]*gatewayv1.Measurement, 0, len(resp.GetMeasurements())),
	}
	for _, m := range resp.GetMeasurements() {
		out.Measurements = append(out.Measurements, measurementFromProfile(m))
	}
	return out, nil
}

func (server *Server) SetMyMeasurementSharing(ctx context.Context, request *gatewayv1.SetMeasurementSharingRequest) (*emptypb.Empty, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	clientUserID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if _, err := server.profileClient.SetMeasurementSharing(forwardContext(ctx), &profilev1.SetMeasurementSharingRequest{
		ClientUserId:   clientUserID,
		TrainerUserIds: request.GetTrainerUserIds(),
	}); err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 204); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (server *Server) GetMyMeasurementSharing(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.MeasurementSharingResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	clientUserID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	resp, err := server.profileClient.GetMeasurementSharing(forwardContext(ctx), &profilev1.GetMeasurementSharingRequest{
		ClientUserId: clientUserID,
	})
	if err != nil {
		return nil, err
	}

	return &gatewayv1.MeasurementSharingResponse{TrainerUserIds: resp.GetTrainerUserIds()}, nil
}

func (server *Server) DeleteMyMeasurement(ctx context.Context, request *gatewayv1.DeleteMeasurementRequest) (*emptypb.Empty, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if _, err := server.profileClient.DeleteMeasurement(forwardContext(ctx), &profilev1.DeleteMeasurementRequest{
		UserId:        userID,
		MeasurementId: request.GetMeasurementId(),
	}); err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 204); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func measurementFromProfile(m *profilev1.Measurement) *gatewayv1.Measurement {
	if m == nil {
		return nil
	}
	return &gatewayv1.Measurement{
		MeasurementId: m.GetMeasurementId(),
		UserId:        m.GetUserId(),
		MeasuredAt:    m.GetMeasuredAt(),
		WeightKg:      m.WeightKg,
		BodyFatPct:    m.BodyFatPct,
		ChestCm:       m.ChestCm,
		WaistCm:       m.WaistCm,
		HipsCm:        m.HipsCm,
		Notes:         m.Notes,
		CreatedAt:     m.GetCreatedAt(),
		UpdatedAt:     m.GetUpdatedAt(),
	}
}

func measurementResponseFromProfile(resp *profilev1.MeasurementResponse) *gatewayv1.MeasurementResponse {
	return &gatewayv1.MeasurementResponse{Measurement: measurementFromProfile(resp.GetMeasurement())}
}
