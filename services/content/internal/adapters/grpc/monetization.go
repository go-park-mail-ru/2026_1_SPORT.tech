package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) ListSubscriptionTiers(ctx context.Context, request *contentv1.ListSubscriptionTiersRequest) (*contentv1.ListSubscriptionTiersResponse, error) {
	tiers, err := server.useCases.Tiers.ListSubscriptionTiers(ctx, mappers.ListSubscriptionTiersRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListSubscriptionTiersResponse(tiers), nil
}

func (server *Server) CreateSubscriptionTier(ctx context.Context, request *contentv1.CreateSubscriptionTierRequest) (*contentv1.SubscriptionTier, error) {
	tier, err := server.useCases.Tiers.CreateSubscriptionTier(ctx, mappers.CreateSubscriptionTierRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSubscriptionTierResponse(tier), nil
}

func (server *Server) UpdateSubscriptionTier(ctx context.Context, request *contentv1.UpdateSubscriptionTierRequest) (*contentv1.SubscriptionTier, error) {
	tier, err := server.useCases.Tiers.UpdateSubscriptionTier(ctx, mappers.UpdateSubscriptionTierRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSubscriptionTierResponse(tier), nil
}

func (server *Server) DeleteSubscriptionTier(ctx context.Context, request *contentv1.DeleteSubscriptionTierRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Tiers.DeleteSubscriptionTier(ctx, mappers.DeleteSubscriptionTierRequestToCommand(request)); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.Empty(), nil
}

func (server *Server) SubscribeToTrainer(ctx context.Context, request *contentv1.SubscribeToTrainerRequest) (*contentv1.Subscription, error) {
	subscription, err := server.useCases.Subscriptions.SubscribeToTrainer(ctx, mappers.SubscribeToTrainerRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSubscriptionResponse(subscription), nil
}

func (server *Server) ListMySubscriptions(ctx context.Context, request *contentv1.ListMySubscriptionsRequest) (*contentv1.ListMySubscriptionsResponse, error) {
	subscriptions, err := server.useCases.Subscriptions.ListMySubscriptions(ctx, mappers.ListMySubscriptionsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListMySubscriptionsResponse(subscriptions), nil
}

func (server *Server) ListTrainerSubscribers(ctx context.Context, request *contentv1.ListTrainerSubscribersRequest) (*contentv1.ListTrainerSubscribersResponse, error) {
	subscribers, err := server.useCases.Subscriptions.ListTrainerSubscribers(ctx, mappers.ListTrainerSubscribersRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListTrainerSubscribersResponse(subscribers), nil
}

func (server *Server) UpdateSubscription(ctx context.Context, request *contentv1.UpdateSubscriptionRequest) (*contentv1.Subscription, error) {
	subscription, err := server.useCases.Subscriptions.UpdateSubscription(ctx, mappers.UpdateSubscriptionRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSubscriptionResponse(subscription), nil
}

func (server *Server) CancelSubscription(ctx context.Context, request *contentv1.CancelSubscriptionRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Subscriptions.CancelSubscription(ctx, mappers.CancelSubscriptionRequestToCommand(request)); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.Empty(), nil
}
