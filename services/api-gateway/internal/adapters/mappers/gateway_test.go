package mappers

import (
	"testing"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAuthMappers(t *testing.T) {
	client := RegisterClientRequestToAuth(&gatewayv1.ClientRegisterRequest{
		Email:    "client@example.com",
		Username: "client",
		Password: "pass1234",
	})
	if client.GetRole() != authv1.UserRole_USER_ROLE_CLIENT || client.GetEmail() != "client@example.com" {
		t.Fatalf("unexpected client register: %+v", client)
	}

	trainer := RegisterTrainerRequestToAuth(&gatewayv1.TrainerRegisterRequest{
		Email:    "trainer@example.com",
		Username: "trainer",
		Password: "pass1234",
	})
	if trainer.GetRole() != authv1.UserRole_USER_ROLE_TRAINER || trainer.GetUsername() != "trainer" {
		t.Fatalf("unexpected trainer register: %+v", trainer)
	}

	login := LoginRequestToAuth(&gatewayv1.LoginRequest{Email: "u@example.com", Password: "pass"})
	if login.GetEmail() != "u@example.com" || login.GetPassword() != "pass" {
		t.Fatalf("unexpected login: %+v", login)
	}

	now := timestamppb.New(time.Date(2026, time.May, 6, 12, 0, 0, 0, time.UTC))
	bio := "bio"
	avatar := "http://cdn/avatar.jpg"
	response, err := AuthResponseFromServices(
		&authv1.AuthUser{
			UserId:   1001,
			Email:    "trainer@example.com",
			Role:     authv1.UserRole_USER_ROLE_ADMIN,
			Status:   authv1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
			Username: "auth-name",
		},
		&profilev1.Profile{
			UserId:    1001,
			Username:  "profile-name",
			FirstName: "Анна",
			LastName:  "Павлова",
			Bio:       &bio,
			AvatarUrl: &avatar,
			IsTrainer: true,
			CreatedAt: now,
			UpdatedAt: now,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetUser().GetUserId() != 1001 ||
		response.GetUser().GetUsername() != "profile-name" ||
		!response.GetUser().GetIsAdmin() ||
		response.GetUser().GetBio() != "bio" ||
		response.GetUser().GetAvatarUrl() != "http://cdn/avatar.jpg" {
		t.Fatalf("unexpected auth response: %+v", response)
	}

	if _, err := UserFromServices(nil, &profilev1.Profile{}); err == nil {
		t.Fatal("expected nil user error")
	}
	if PasswordsMatch("a", "b") {
		t.Fatal("passwords should not match")
	}
	if err := RequireTrainerRole(&authv1.AuthUser{Role: authv1.UserRole_USER_ROLE_CLIENT}); err == nil {
		t.Fatal("expected trainer role error")
	}
	if err := RequireTrainerRole(&authv1.AuthUser{Role: authv1.UserRole_USER_ROLE_TRAINER}); err != nil {
		t.Fatalf("unexpected trainer role error: %v", err)
	}
	if NormalizeStatusCode(" BAD_REQUEST ") != "bad_request" || NormalizeStatusCode("custom") != "custom" {
		t.Fatal("unexpected status normalization")
	}
}

func TestDonationMappers(t *testing.T) {
	message := "Спасибо"
	request := DonateToProfileRequestToContent(1002, &gatewayv1.DonateToProfileRequest{
		UserId:      1001,
		AmountValue: 1500,
		Currency:    "RUB",
		Message:     &message,
	})
	if request.GetSenderUserId() != 1002 || request.GetRecipientUserId() != 1001 || request.GetMessage() != "Спасибо" {
		t.Fatalf("unexpected donation request: %+v", request)
	}

	now := timestamppb.Now()
	response, err := DonationResponseFromContent(&contentv1.DonationResponse{
		Donation: &contentv1.Donation{
			DonationId:      77,
			SenderUserId:    1002,
			RecipientUserId: 1001,
			AmountValue:     1500,
			Currency:        "RUB",
			Message:         &message,
			CreatedAt:       now,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetDonationId() != 77 || response.GetSenderUserId() != 1002 || response.GetMessage() != "Спасибо" {
		t.Fatalf("unexpected donation response: %+v", response)
	}
	if _, err := DonationResponseFromContent(nil); err == nil {
		t.Fatal("expected nil donation error")
	}

	balance, err := BalanceResponseFromContent(&contentv1.BalanceResponse{
		TrainerUserId: 1001,
		AmountValue:   2500,
		Currency:      "RUB",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance.GetTrainerId() != 1001 || balance.GetAmountValue() != 2500 {
		t.Fatalf("unexpected balance response: %+v", balance)
	}
}

func TestSubscriptionMappers(t *testing.T) {
	subscribe := SubscribeRequestToContent(1002, &gatewayv1.SubscribeRequest{TrainerId: 1001, TierId: 2})
	if subscribe.GetClientUserId() != 1002 || subscribe.GetTrainerUserId() != 1001 || subscribe.GetTierId() != 2 {
		t.Fatalf("unexpected subscribe request: %+v", subscribe)
	}
	update := UpdateSubscriptionRequestToContent(1002, &gatewayv1.UpdateSubscriptionRequest{SubscriptionId: 55, TierId: 3})
	if update.GetClientUserId() != 1002 || update.GetSubscriptionId() != 55 || update.GetTierId() != 3 {
		t.Fatalf("unexpected update request: %+v", update)
	}

	now := timestamppb.Now()
	subscription, err := SubscriptionFromContent(&contentv1.Subscription{
		SubscriptionId: 55,
		TrainerUserId:  1001,
		TierId:         3,
		TierName:       "Premium",
		Price:          1500,
		Active:         true,
		ExpiresAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subscription.GetSubscriptionId() != 55 || subscription.GetTrainerId() != 1001 || !subscription.GetActive() {
		t.Fatalf("unexpected subscription: %+v", subscription)
	}
	if _, err := SubscriptionFromContent(nil); err == nil {
		t.Fatal("expected nil subscription error")
	}

	list, err := SubscriptionsResponseFromContent(&contentv1.ListMySubscriptionsResponse{
		Subscriptions: []*contentv1.Subscription{{SubscriptionId: 1, TrainerUserId: 2, TierId: 3}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.GetSubscriptions()) != 1 {
		t.Fatalf("unexpected subscriptions response: %+v", list)
	}
	empty, err := SubscriptionsResponseFromContent(nil)
	if err != nil || len(empty.GetSubscriptions()) != 0 {
		t.Fatalf("unexpected empty subscriptions: %+v err=%v", empty, err)
	}
}

func TestTierMappers(t *testing.T) {
	description := "desc"
	create := CreateTierRequestToContent(1001, &gatewayv1.CreateTierRequest{
		Name:        "Basic",
		Price:       500,
		Description: &description,
	})
	if create.GetTrainerUserId() != 1001 || create.GetName() != "Basic" || create.GetDescription() != "desc" {
		t.Fatalf("unexpected create tier request: %+v", create)
	}

	price := int32(700)
	update := UpdateTierRequestToContent(1001, &gatewayv1.UpdateTierRequest{
		TierId:           2,
		Name:             stringPtr("Advanced"),
		Price:            &price,
		Description:      &description,
		ClearDescription: true,
	})
	if update.GetTrainerUserId() != 1001 || update.GetTierId() != 2 || update.GetPrice() != 700 || !update.GetClearDescription() {
		t.Fatalf("unexpected update tier request: %+v", update)
	}

	now := timestamppb.Now()
	tier, err := TierFromContent(&contentv1.SubscriptionTier{
		TierId:      2,
		Name:        "Advanced",
		Price:       700,
		Description: &description,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier.GetTierId() != 2 || tier.GetDescription() != "desc" {
		t.Fatalf("unexpected tier: %+v", tier)
	}
	if _, err := TierFromContent(nil); err == nil {
		t.Fatal("expected nil tier error")
	}

	tiers, err := TiersResponseFromContent(&contentv1.ListSubscriptionTiersResponse{
		Tiers: []*contentv1.SubscriptionTier{{TierId: 1, Name: "Basic"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tiers.GetTiers()) != 1 {
		t.Fatalf("unexpected tiers: %+v", tiers)
	}
}
