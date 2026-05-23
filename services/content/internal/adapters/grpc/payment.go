package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
)

func (server *Server) DonateToProfile(ctx context.Context, request *contentv1.DonateToProfileRequest) (*contentv1.DonationResponse, error) {
	donation, err := server.useCases.Donations.DonateToProfile(ctx, mappers.DonateToProfileRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewDonationResponse(donation), nil
}

func (server *Server) CreateDonationPayment(ctx context.Context, request *contentv1.CreateDonationPaymentRequest) (*contentv1.PaymentResponse, error) {
	payment, err := server.useCases.Donations.CreateDonationPayment(ctx, mappers.CreateDonationPaymentRequestToCommand(request))
	if err != nil {
		return nil, server.statusError("CreateDonationPayment", err)
	}

	return mappers.NewPaymentResponse(payment), nil
}

func (server *Server) CreateSubscriptionPayment(ctx context.Context, request *contentv1.CreateSubscriptionPaymentRequest) (*contentv1.PaymentResponse, error) {
	payment, err := server.useCases.Donations.CreateSubscriptionPayment(ctx, mappers.CreateSubscriptionPaymentRequestToCommand(request))
	if err != nil {
		return nil, server.statusError("CreateSubscriptionPayment", err)
	}

	return mappers.NewPaymentResponse(payment), nil
}

func (server *Server) ConfirmDonationPayment(ctx context.Context, request *contentv1.ConfirmDonationPaymentRequest) (*contentv1.PaymentResponse, error) {
	payment, err := server.useCases.Donations.ConfirmDonationPayment(ctx, mappers.ConfirmDonationPaymentRequestToCommand(request))
	if err != nil {
		return nil, server.statusError("ConfirmDonationPayment", err)
	}

	return mappers.NewPaymentResponse(payment), nil
}

func (server *Server) GetBalance(ctx context.Context, request *contentv1.GetBalanceRequest) (*contentv1.BalanceResponse, error) {
	balance, err := server.useCases.Donations.GetBalance(ctx, mappers.GetBalanceRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewBalanceResponse(balance), nil
}

func (server *Server) GetTrainerStatistics(ctx context.Context, request *contentv1.GetTrainerStatisticsRequest) (*contentv1.TrainerStatisticsResponse, error) {
	statistics, err := server.useCases.Donations.GetTrainerStatistics(ctx, mappers.GetTrainerStatisticsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewTrainerStatisticsResponse(statistics), nil
}

func (server *Server) ListReceivedDonations(ctx context.Context, request *contentv1.ListReceivedDonationsRequest) (*contentv1.ListReceivedDonationsResponse, error) {
	donations, total, err := server.useCases.Donations.ListReceivedDonations(ctx, usecase.ListReceivedDonationsQuery{
		TrainerUserID: request.GetTrainerUserId(),
		Limit:         request.GetLimit(),
		Offset:        request.GetOffset(),
	})
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	records := make([]*contentv1.DonationRecord, 0, len(donations))
	for _, d := range donations {
		rec := &contentv1.DonationRecord{
			DonationId:   d.DonationID,
			SenderUserId: d.SenderUserID,
			AmountValue:  d.AmountValue,
			Currency:     d.Currency,
			CreatedAt:    d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if d.Message != nil {
			rec.Message = d.Message
		}
		records = append(records, rec)
	}

	return &contentv1.ListReceivedDonationsResponse{
		Donations: records,
		Total:     total,
	}, nil
}
