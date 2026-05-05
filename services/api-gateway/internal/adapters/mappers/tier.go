package mappers

import (
	"fmt"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func CreateTierRequestToContent(trainerUserID int64, request *gatewayv1.CreateTierRequest) *contentv1.CreateSubscriptionTierRequest {
	return &contentv1.CreateSubscriptionTierRequest{
		TrainerUserId: trainerUserID,
		Name:          request.GetName(),
		Price:         request.GetPrice(),
		Description:   request.Description,
	}
}

func UpdateTierRequestToContent(trainerUserID int64, request *gatewayv1.UpdateTierRequest) *contentv1.UpdateSubscriptionTierRequest {
	return &contentv1.UpdateSubscriptionTierRequest{
		TrainerUserId:    trainerUserID,
		TierId:           int32ToInt64(request.GetTierId()),
		Name:             request.Name,
		Price:            request.Price,
		Description:      request.Description,
		ClearDescription: request.GetClearDescription(),
	}
}

func TiersResponseFromContent(response *contentv1.ListSubscriptionTiersResponse) (*gatewayv1.TiersResponse, error) {
	tiers := make([]*gatewayv1.Tier, 0)
	if response != nil {
		tiers = make([]*gatewayv1.Tier, 0, len(response.GetTiers()))
		for _, tier := range response.GetTiers() {
			mappedTier, err := TierFromContent(tier)
			if err != nil {
				return nil, err
			}
			tiers = append(tiers, mappedTier)
		}
	}

	return &gatewayv1.TiersResponse{Tiers: tiers}, nil
}

func TierFromContent(tier *contentv1.SubscriptionTier) (*gatewayv1.Tier, error) {
	if tier == nil {
		return nil, fmt.Errorf("subscription tier is required")
	}

	tierID, err := int64ToInt32("content.subscription_tier.tier_id", tier.GetTierId())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.Tier{
		TierId:      tierID,
		Name:        tier.GetName(),
		Price:       tier.GetPrice(),
		Description: tier.Description,
		CreatedAt:   tier.GetCreatedAt(),
		UpdatedAt:   tier.GetUpdatedAt(),
	}, nil
}
