package mappers

import (
	"fmt"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func SubscribeRequestToContent(clientUserID int64, request *gatewayv1.SubscribeRequest) *contentv1.SubscribeToTrainerRequest {
	return &contentv1.SubscribeToTrainerRequest{
		ClientUserId:  clientUserID,
		TrainerUserId: int32ToInt64(request.GetTrainerId()),
		TierId:        int32ToInt64(request.GetTierId()),
	}
}

func UpdateSubscriptionRequestToContent(clientUserID int64, request *gatewayv1.UpdateSubscriptionRequest) *contentv1.UpdateSubscriptionRequest {
	return &contentv1.UpdateSubscriptionRequest{
		ClientUserId:   clientUserID,
		SubscriptionId: int32ToInt64(request.GetSubscriptionId()),
		TierId:         int32ToInt64(request.GetTierId()),
	}
}

func SubscriptionFromContent(subscription *contentv1.Subscription) (*gatewayv1.Subscription, error) {
	if subscription == nil {
		return nil, fmt.Errorf("subscription is required")
	}

	subscriptionID, err := int64ToInt32("content.subscription.subscription_id", subscription.GetSubscriptionId())
	if err != nil {
		return nil, err
	}
	trainerID, err := int64ToInt32("content.subscription.trainer_user_id", subscription.GetTrainerUserId())
	if err != nil {
		return nil, err
	}
	tierID, err := int64ToInt32("content.subscription.tier_id", subscription.GetTierId())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.Subscription{
		SubscriptionId: subscriptionID,
		TrainerId:      trainerID,
		TierId:         tierID,
		TierName:       subscription.GetTierName(),
		Price:          subscription.GetPrice(),
		Active:         subscription.GetActive(),
		ExpiresAt:      subscription.GetExpiresAt(),
		CreatedAt:      subscription.GetCreatedAt(),
		UpdatedAt:      subscription.GetUpdatedAt(),
	}, nil
}

func SubscriptionsResponseFromContent(response *contentv1.ListMySubscriptionsResponse) (*gatewayv1.SubscriptionsResponse, error) {
	subscriptions := make([]*gatewayv1.Subscription, 0)
	if response != nil {
		subscriptions = make([]*gatewayv1.Subscription, 0, len(response.GetSubscriptions()))
		for _, subscription := range response.GetSubscriptions() {
			mappedSubscription, err := SubscriptionFromContent(subscription)
			if err != nil {
				return nil, err
			}

			subscriptions = append(subscriptions, mappedSubscription)
		}
	}

	return &gatewayv1.SubscriptionsResponse{Subscriptions: subscriptions}, nil
}
