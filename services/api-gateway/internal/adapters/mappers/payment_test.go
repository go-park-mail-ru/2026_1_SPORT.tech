package mappers

import (
	"testing"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func TestCreateDonationPaymentRequestToContent(t *testing.T) {
	msg := "great work"
	retURL := "https://example.com/return"
	cancelURL := "https://example.com/cancel"
	req := CreateDonationPaymentRequestToContent(5, &gatewayv1.CreateDonationPaymentRequest{
		UserId:      42,
		AmountValue: 500,
		Currency:    "RUB",
		Message:     &msg,
		ReturnUrl:   &retURL,
		CancelUrl:   &cancelURL,
	})
	if req.GetSenderUserId() != 5 || req.GetRecipientUserId() != 42 {
		t.Fatalf("unexpected sender/recipient: %d / %d", req.GetSenderUserId(), req.GetRecipientUserId())
	}
	if req.GetAmountValue() != 500 || req.GetCurrency() != "RUB" {
		t.Fatalf("unexpected amount/currency: %v / %s", req.GetAmountValue(), req.GetCurrency())
	}
}

func TestCreateSubscriptionPaymentRequestToContent(t *testing.T) {
	retURL := "https://example.com/return"
	cancelURL := "https://example.com/cancel"
	req := CreateSubscriptionPaymentRequestToContent(3, &gatewayv1.CreateSubscriptionPaymentRequest{
		TrainerId: 10,
		TierId:    2,
		ReturnUrl: &retURL,
		CancelUrl: &cancelURL,
	})
	if req.GetClientUserId() != 3 || req.GetTrainerUserId() != 10 || req.GetTierId() != 2 {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestConfirmDonationPaymentRequestToContent(t *testing.T) {
	req := ConfirmDonationPaymentRequestToContent(7, &gatewayv1.ConfirmDonationPaymentRequest{
		PaymentId:         99,
		ConfirmationToken: "tok_abc",
	})
	if req.GetSenderUserId() != 7 || req.GetPaymentId() != 99 || req.GetConfirmationToken() != "tok_abc" {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestPaymentResponseFromContent(t *testing.T) {
	resp, err := PaymentResponseFromContent(&contentv1.PaymentResponse{
		Payment: &contentv1.Payment{
			PaymentId:       1,
			SenderUserId:    2,
			RecipientUserId: 3,
			AmountValue:     1000,
			Currency:        "RUB",
			Status:          "pending",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetPaymentId() != 1 || resp.GetSenderUserId() != 2 || resp.GetRecipientUserId() != 3 {
		t.Fatalf("unexpected payment: %+v", resp)
	}
	if resp.GetAmountValue() != 1000 || resp.GetCurrency() != "RUB" {
		t.Fatalf("unexpected amount: %v %s", resp.GetAmountValue(), resp.GetCurrency())
	}
}

func TestPaymentResponseFromContentNil(t *testing.T) {
	_, err := PaymentResponseFromContent(nil)
	if err == nil {
		t.Fatal("expected error for nil response")
	}

	_, err = PaymentResponseFromContent(&contentv1.PaymentResponse{})
	if err == nil {
		t.Fatal("expected error for nil payment")
	}
}

func TestPaymentResponseFromContentWithDonation(t *testing.T) {
	resp, err := PaymentResponseFromContent(&contentv1.PaymentResponse{
		Payment: &contentv1.Payment{
			PaymentId:       1,
			SenderUserId:    2,
			RecipientUserId: 3,
			Status:          "pending",
			Donation: &contentv1.Donation{
				DonationId:      10,
				SenderUserId:    2,
				RecipientUserId: 3,
				AmountValue:     500,
				Currency:        "RUB",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetDonation() == nil {
		t.Fatal("expected donation to be set")
	}
	if resp.GetDonation().GetDonationId() != 10 {
		t.Fatalf("unexpected donation id: %d", resp.GetDonation().GetDonationId())
	}
}
