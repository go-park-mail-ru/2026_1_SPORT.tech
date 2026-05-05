package mappers

import (
	"fmt"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func DonateToProfileRequestToContent(senderUserID int64, request *gatewayv1.DonateToProfileRequest) *contentv1.DonateToProfileRequest {
	return &contentv1.DonateToProfileRequest{
		SenderUserId:    senderUserID,
		RecipientUserId: int32ToInt64(request.GetUserId()),
		AmountValue:     request.GetAmountValue(),
		Currency:        request.GetCurrency(),
		Message:         request.Message,
	}
}

func DonationResponseFromContent(response *contentv1.DonationResponse) (*gatewayv1.DonationResponse, error) {
	if response == nil || response.GetDonation() == nil {
		return nil, fmt.Errorf("donation is required")
	}

	donation := response.GetDonation()
	donationID, err := int64ToInt32("content.donation.donation_id", donation.GetDonationId())
	if err != nil {
		return nil, err
	}
	senderUserID, err := int64ToInt32("content.donation.sender_user_id", donation.GetSenderUserId())
	if err != nil {
		return nil, err
	}
	recipientUserID, err := int64ToInt32("content.donation.recipient_user_id", donation.GetRecipientUserId())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.DonationResponse{
		DonationId:      donationID,
		SenderUserId:    senderUserID,
		RecipientUserId: recipientUserID,
		AmountValue:     donation.GetAmountValue(),
		Currency:        donation.GetCurrency(),
		Message:         donation.Message,
		CreatedAt:       donation.GetCreatedAt(),
	}, nil
}

func BalanceResponseFromContent(response *contentv1.BalanceResponse) (*gatewayv1.BalanceResponse, error) {
	if response == nil {
		return nil, fmt.Errorf("balance is required")
	}

	trainerID, err := int64ToInt32("content.balance.trainer_user_id", response.GetTrainerUserId())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.BalanceResponse{
		TrainerId:   trainerID,
		AmountValue: response.GetAmountValue(),
		Currency:    response.GetCurrency(),
	}, nil
}
