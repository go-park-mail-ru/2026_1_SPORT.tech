package mappers

import (
	"fmt"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func CreateDonationPaymentRequestToContent(senderUserID int64, request *gatewayv1.CreateDonationPaymentRequest) *contentv1.CreateDonationPaymentRequest {
	return &contentv1.CreateDonationPaymentRequest{
		SenderUserId:    senderUserID,
		RecipientUserId: int32ToInt64(request.GetUserId()),
		AmountValue:     request.GetAmountValue(),
		Currency:        request.GetCurrency(),
		Message:         request.Message,
		ReturnUrl:       request.ReturnUrl,
		CancelUrl:       request.CancelUrl,
	}
}

func CreateSubscriptionPaymentRequestToContent(clientUserID int64, request *gatewayv1.CreateSubscriptionPaymentRequest) *contentv1.CreateSubscriptionPaymentRequest {
	return &contentv1.CreateSubscriptionPaymentRequest{
		ClientUserId:  clientUserID,
		TrainerUserId: int32ToInt64(request.GetTrainerId()),
		TierId:        int32ToInt64(request.GetTierId()),
		ReturnUrl:     request.ReturnUrl,
		CancelUrl:     request.CancelUrl,
	}
}

func ConfirmDonationPaymentRequestToContent(senderUserID int64, request *gatewayv1.ConfirmDonationPaymentRequest) *contentv1.ConfirmDonationPaymentRequest {
	return &contentv1.ConfirmDonationPaymentRequest{
		SenderUserId:      senderUserID,
		PaymentId:         int32ToInt64(request.GetPaymentId()),
		ConfirmationToken: request.GetConfirmationToken(),
	}
}

func PaymentResponseFromContent(response *contentv1.PaymentResponse) (*gatewayv1.PaymentResponse, error) {
	if response == nil || response.GetPayment() == nil {
		return nil, fmt.Errorf("payment is required")
	}

	payment := response.GetPayment()
	paymentID, err := int64ToInt32("content.payment.payment_id", payment.GetPaymentId())
	if err != nil {
		return nil, err
	}
	senderUserID, err := int64ToInt32("content.payment.sender_user_id", payment.GetSenderUserId())
	if err != nil {
		return nil, err
	}
	recipientUserID, err := int64ToInt32("content.payment.recipient_user_id", payment.GetRecipientUserId())
	if err != nil {
		return nil, err
	}

	result := &gatewayv1.PaymentResponse{
		PaymentId:         paymentID,
		ProviderPaymentId: payment.GetProviderPaymentId(),
		Status:            payment.GetStatus(),
		SenderUserId:      senderUserID,
		RecipientUserId:   recipientUserID,
		AmountValue:       payment.GetAmountValue(),
		Currency:          payment.GetCurrency(),
		Message:           payment.Message,
		ConfirmationToken: payment.GetConfirmationToken(),
		ConfirmationUrl:   payment.GetConfirmationUrl(),
		CreatedAt:         payment.GetCreatedAt(),
		UpdatedAt:         payment.GetUpdatedAt(),
		ConfirmedAt:       payment.ConfirmedAt,
	}
	if payment.GetDonation() != nil {
		donation, err := paymentDonationFromContent(payment.GetDonation())
		if err != nil {
			return nil, err
		}
		result.Donation = donation
	}
	if payment.GetSubscription() != nil {
		subscription, err := SubscriptionFromContent(payment.GetSubscription())
		if err != nil {
			return nil, err
		}
		result.Subscription = subscription
	}

	return result, nil
}

func paymentDonationFromContent(donation *contentv1.Donation) (*gatewayv1.PaymentDonation, error) {
	donationID, err := int64ToInt32("content.payment.donation.donation_id", donation.GetDonationId())
	if err != nil {
		return nil, err
	}
	senderUserID, err := int64ToInt32("content.payment.donation.sender_user_id", donation.GetSenderUserId())
	if err != nil {
		return nil, err
	}
	recipientUserID, err := int64ToInt32("content.payment.donation.recipient_user_id", donation.GetRecipientUserId())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PaymentDonation{
		DonationId:      donationID,
		SenderUserId:    senderUserID,
		RecipientUserId: recipientUserID,
		AmountValue:     donation.GetAmountValue(),
		Currency:        donation.GetCurrency(),
		Message:         donation.Message,
		CreatedAt:       donation.GetCreatedAt(),
	}, nil
}
