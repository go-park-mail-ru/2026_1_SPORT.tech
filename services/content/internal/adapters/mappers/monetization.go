package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ListSubscriptionTiersRequestToQuery(request *contentv1.ListSubscriptionTiersRequest) usecase.ListSubscriptionTiersQuery {
	return usecase.ListSubscriptionTiersQuery{
		TrainerUserID: request.GetTrainerUserId(),
	}
}

func CreateSubscriptionTierRequestToCommand(request *contentv1.CreateSubscriptionTierRequest) usecase.CreateSubscriptionTierCommand {
	return usecase.CreateSubscriptionTierCommand{
		TrainerUserID: request.GetTrainerUserId(),
		Name:          request.GetName(),
		Price:         request.GetPrice(),
		Description:   request.Description,
		ChatEnabled:   request.GetChatEnabled(),
	}
}

func UpdateSubscriptionTierRequestToCommand(request *contentv1.UpdateSubscriptionTierRequest) usecase.UpdateSubscriptionTierCommand {
	return usecase.UpdateSubscriptionTierCommand{
		TrainerUserID:    request.GetTrainerUserId(),
		TierID:           request.GetTierId(),
		Name:             request.Name,
		Price:            request.Price,
		Description:      request.Description,
		ClearDescription: request.GetClearDescription(),
		ChatEnabled:      request.ChatEnabled,
	}
}

func DeleteSubscriptionTierRequestToCommand(request *contentv1.DeleteSubscriptionTierRequest) usecase.DeleteSubscriptionTierCommand {
	return usecase.DeleteSubscriptionTierCommand{
		TrainerUserID: request.GetTrainerUserId(),
		TierID:        request.GetTierId(),
	}
}

func SubscribeToTrainerRequestToCommand(request *contentv1.SubscribeToTrainerRequest) usecase.SubscribeToTrainerCommand {
	return usecase.SubscribeToTrainerCommand{
		ClientUserID:  request.GetClientUserId(),
		TrainerUserID: request.GetTrainerUserId(),
		TierID:        request.GetTierId(),
	}
}

func ListMySubscriptionsRequestToQuery(request *contentv1.ListMySubscriptionsRequest) usecase.ListMySubscriptionsQuery {
	return usecase.ListMySubscriptionsQuery{
		ClientUserID: request.GetClientUserId(),
	}
}

func ListTrainerSubscribersRequestToQuery(request *contentv1.ListTrainerSubscribersRequest) usecase.ListTrainerSubscribersQuery {
	return usecase.ListTrainerSubscribersQuery{
		TrainerUserID: request.GetTrainerUserId(),
		Limit:         request.GetLimit(),
		Offset:        request.GetOffset(),
	}
}

func UpdateSubscriptionRequestToCommand(request *contentv1.UpdateSubscriptionRequest) usecase.UpdateSubscriptionCommand {
	return usecase.UpdateSubscriptionCommand{
		ClientUserID:   request.GetClientUserId(),
		SubscriptionID: request.GetSubscriptionId(),
		TierID:         request.GetTierId(),
	}
}

func CancelSubscriptionRequestToCommand(request *contentv1.CancelSubscriptionRequest) usecase.CancelSubscriptionCommand {
	return usecase.CancelSubscriptionCommand{
		ClientUserID:   request.GetClientUserId(),
		SubscriptionID: request.GetSubscriptionId(),
	}
}

func DonateToProfileRequestToCommand(request *contentv1.DonateToProfileRequest) usecase.DonateToProfileCommand {
	return usecase.DonateToProfileCommand{
		SenderUserID:    request.GetSenderUserId(),
		RecipientUserID: request.GetRecipientUserId(),
		AmountValue:     request.GetAmountValue(),
		Currency:        request.GetCurrency(),
		Message:         request.Message,
	}
}

func CreateDonationPaymentRequestToCommand(request *contentv1.CreateDonationPaymentRequest) usecase.CreateDonationPaymentCommand {
	return usecase.CreateDonationPaymentCommand{
		SenderUserID:    request.GetSenderUserId(),
		RecipientUserID: request.GetRecipientUserId(),
		AmountValue:     request.GetAmountValue(),
		Currency:        request.GetCurrency(),
		Message:         request.Message,
		ReturnURL:       request.ReturnUrl,
		CancelURL:       request.CancelUrl,
	}
}

func CreateSubscriptionPaymentRequestToCommand(request *contentv1.CreateSubscriptionPaymentRequest) usecase.CreateSubscriptionPaymentCommand {
	return usecase.CreateSubscriptionPaymentCommand{
		ClientUserID:  request.GetClientUserId(),
		TrainerUserID: request.GetTrainerUserId(),
		TierID:        request.GetTierId(),
		ReturnURL:     request.ReturnUrl,
		CancelURL:     request.CancelUrl,
	}
}

func ConfirmDonationPaymentRequestToCommand(request *contentv1.ConfirmDonationPaymentRequest) usecase.ConfirmDonationPaymentCommand {
	return usecase.ConfirmDonationPaymentCommand{
		SenderUserID:      request.GetSenderUserId(),
		PaymentID:         request.GetPaymentId(),
		ConfirmationToken: request.GetConfirmationToken(),
	}
}

func GetBalanceRequestToQuery(request *contentv1.GetBalanceRequest) usecase.GetBalanceQuery {
	return usecase.GetBalanceQuery{
		TrainerUserID: request.GetTrainerUserId(),
		Currency:      request.GetCurrency(),
	}
}

func GetTrainerStatisticsRequestToQuery(request *contentv1.GetTrainerStatisticsRequest) usecase.GetTrainerStatisticsQuery {
	return usecase.GetTrainerStatisticsQuery{
		TrainerUserID: request.GetTrainerUserId(),
		Currency:      request.GetCurrency(),
	}
}

func NewListSubscriptionTiersResponse(tiers []domain.SubscriptionTier) *contentv1.ListSubscriptionTiersResponse {
	response := &contentv1.ListSubscriptionTiersResponse{
		Tiers: make([]*contentv1.SubscriptionTier, 0, len(tiers)),
	}
	for _, tier := range tiers {
		response.Tiers = append(response.Tiers, subscriptionTierToProto(tier))
	}

	return response
}

func NewSubscriptionTierResponse(tier domain.SubscriptionTier) *contentv1.SubscriptionTier {
	return subscriptionTierToProto(tier)
}

func NewSubscriptionResponse(subscription domain.Subscription) *contentv1.Subscription {
	return subscriptionToProto(subscription)
}

func NewListMySubscriptionsResponse(subscriptions []domain.Subscription) *contentv1.ListMySubscriptionsResponse {
	response := &contentv1.ListMySubscriptionsResponse{
		Subscriptions: make([]*contentv1.Subscription, 0, len(subscriptions)),
	}
	for _, subscription := range subscriptions {
		response.Subscriptions = append(response.Subscriptions, subscriptionToProto(subscription))
	}

	return response
}

func NewListTrainerSubscribersResponse(subscribers []domain.Subscription) *contentv1.ListTrainerSubscribersResponse {
	response := &contentv1.ListTrainerSubscribersResponse{
		Subscribers: make([]*contentv1.Subscription, 0, len(subscribers)),
	}
	for _, subscriber := range subscribers {
		response.Subscribers = append(response.Subscribers, subscriptionToProto(subscriber))
	}

	return response
}

func NewDonationResponse(donation domain.Donation) *contentv1.DonationResponse {
	return &contentv1.DonationResponse{
		Donation: donationToProto(donation),
	}
}

func NewPaymentResponse(payment domain.DonationPayment) *contentv1.PaymentResponse {
	return &contentv1.PaymentResponse{
		Payment: paymentToProto(payment),
	}
}

func NewBalanceResponse(balance domain.Balance) *contentv1.BalanceResponse {
	return &contentv1.BalanceResponse{
		TrainerUserId: balance.TrainerUserID,
		AmountValue:   balance.AmountValue,
		Currency:      balance.Currency,
	}
}

func NewTrainerStatisticsResponse(statistics domain.TrainerStatistics) *contentv1.TrainerStatisticsResponse {
	return &contentv1.TrainerStatisticsResponse{
		TrainerUserId:  statistics.TrainerUserID,
		PostsCount:     statistics.PostsCount,
		DonationsCount: statistics.DonationsCount,
		TotalRevenue:   statistics.TotalRevenue,
		MonthlyRevenue: statistics.MonthlyRevenue,
		Currency:       statistics.Currency,
	}
}

func subscriptionTierToProto(tier domain.SubscriptionTier) *contentv1.SubscriptionTier {
	response := &contentv1.SubscriptionTier{
		TierId:        tier.TierID,
		TrainerUserId: tier.TrainerUserID,
		Name:          tier.Name,
		Price:         tier.Price,
		ChatEnabled:   tier.ChatEnabled,
		CreatedAt:     timestamppb.New(tier.CreatedAt),
		UpdatedAt:     timestamppb.New(tier.UpdatedAt),
	}
	if tier.Description != nil {
		response.Description = tier.Description
	}

	return response
}

func subscriptionToProto(subscription domain.Subscription) *contentv1.Subscription {
	return &contentv1.Subscription{
		SubscriptionId: subscription.SubscriptionID,
		ClientUserId:   subscription.ClientUserID,
		TrainerUserId:  subscription.TrainerUserID,
		TierId:         subscription.TierID,
		TierName:       subscription.TierName,
		Price:          subscription.Price,
		Active:         subscription.Active,
		ExpiresAt:      timestamppb.New(subscription.ExpiresAt),
		CreatedAt:      timestamppb.New(subscription.CreatedAt),
		UpdatedAt:      timestamppb.New(subscription.UpdatedAt),
	}
}

func donationToProto(donation domain.Donation) *contentv1.Donation {
	response := &contentv1.Donation{
		DonationId:      donation.DonationID,
		SenderUserId:    donation.SenderUserID,
		RecipientUserId: donation.RecipientUserID,
		AmountValue:     donation.AmountValue,
		Currency:        donation.Currency,
		CreatedAt:       timestamppb.New(donation.CreatedAt),
	}
	if donation.Message != nil {
		response.Message = donation.Message
	}

	return response
}

func paymentToProto(payment domain.DonationPayment) *contentv1.Payment {
	response := &contentv1.Payment{
		PaymentId:         payment.PaymentID,
		ProviderPaymentId: payment.ProviderPaymentID,
		Status:            string(payment.Status),
		SenderUserId:      payment.SenderUserID,
		RecipientUserId:   payment.RecipientUserID,
		AmountValue:       payment.AmountValue,
		Currency:          payment.Currency,
		ConfirmationToken: payment.ConfirmationToken,
		ConfirmationUrl:   payment.ConfirmationURL,
		CreatedAt:         timestamppb.New(payment.CreatedAt),
		UpdatedAt:         timestamppb.New(payment.UpdatedAt),
	}
	if payment.Message != nil {
		response.Message = payment.Message
	}
	if payment.Donation != nil {
		response.Donation = donationToProto(*payment.Donation)
	}
	if payment.Subscription != nil {
		response.Subscription = subscriptionToProto(*payment.Subscription)
	}
	if payment.ConfirmedAt != nil {
		response.ConfirmedAt = timestamppb.New(*payment.ConfirmedAt)
	}

	return response
}
