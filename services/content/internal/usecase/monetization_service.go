package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func (service *Service) ListSubscriptionTiers(ctx context.Context, query ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error) {
	if query.TrainerUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	return service.money.ListSubscriptionTiers(ctx, query.TrainerUserID)
}

func (service *Service) CreateSubscriptionTier(ctx context.Context, command CreateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	tier := domain.SubscriptionTier{
		TrainerUserID: command.TrainerUserID,
		Name:          normalizeRequiredText(command.Name),
		Price:         command.Price,
		Description:   normalizeOptionalText(command.Description),
		ChatEnabled:   command.ChatEnabled,
	}
	if err := validateSubscriptionTier(tier); err != nil {
		return domain.SubscriptionTier{}, err
	}

	return service.money.CreateSubscriptionTier(ctx, tier)
}

func (service *Service) UpdateSubscriptionTier(ctx context.Context, command UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	if err := validateUpdateSubscriptionTierCommand(command); err != nil {
		return domain.SubscriptionTier{}, err
	}

	tier, err := service.money.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
	if err != nil {
		return domain.SubscriptionTier{}, err
	}

	if command.Name != nil {
		tier.Name = normalizeRequiredText(*command.Name)
	}
	if command.Price != nil {
		tier.Price = *command.Price
	}
	switch {
	case command.ClearDescription:
		tier.Description = nil
	case command.Description != nil:
		tier.Description = normalizeOptionalText(command.Description)
	}
	if command.ChatEnabled != nil {
		tier.ChatEnabled = *command.ChatEnabled
	}

	if err := validateSubscriptionTier(tier); err != nil {
		return domain.SubscriptionTier{}, err
	}

	return service.money.UpdateSubscriptionTier(ctx, tier)
}

func (service *Service) DeleteSubscriptionTier(ctx context.Context, command DeleteSubscriptionTierCommand) error {
	if err := validateSubscriptionTierIDCommand(command.TrainerUserID, command.TierID); err != nil {
		return err
	}

	return service.money.DeleteSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
}

func (service *Service) SubscribeToTrainer(ctx context.Context, command SubscribeToTrainerCommand) (domain.Subscription, error) {
	return domain.Subscription{}, ErrSubscriptionPaymentRequired
}

func (service *Service) createPaidSubscription(ctx context.Context, command SubscribeToTrainerCommand) (domain.Subscription, error) {
	if err := validateSubscribeToTrainerCommand(command); err != nil {
		return domain.Subscription{}, err
	}

	tier, err := service.money.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
	if err != nil {
		return domain.Subscription{}, err
	}

	subscription, err := service.money.SubscribeToTrainer(ctx, domain.Subscription{
		ClientUserID:  command.ClientUserID,
		TrainerUserID: command.TrainerUserID,
		TierID:        tier.TierID,
		ExpiresAt:     time.Now().UTC().AddDate(0, 1, 0),
	})
	if err != nil {
		return domain.Subscription{}, err
	}

	if err := service.createSubscriptionNotifications(ctx, subscription); err != nil {
		return domain.Subscription{}, err
	}

	return subscription, nil
}

func (service *Service) ListMySubscriptions(ctx context.Context, query ListMySubscriptionsQuery) ([]domain.Subscription, error) {
	if query.ClientUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	return service.money.ListSubscriptions(ctx, query.ClientUserID)
}

func (service *Service) ListTrainerSubscribers(ctx context.Context, query ListTrainerSubscribersQuery) ([]domain.Subscription, error) {
	if query.TrainerUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return service.money.ListTrainerSubscribers(ctx, query.TrainerUserID, limit, offset)
}

func (service *Service) UpdateSubscription(ctx context.Context, command UpdateSubscriptionCommand) (domain.Subscription, error) {
	return domain.Subscription{}, ErrSubscriptionPaymentRequired
}

func (service *Service) updatePaidSubscription(ctx context.Context, command UpdateSubscriptionCommand) (domain.Subscription, error) {
	if err := validateUpdateSubscriptionCommand(command); err != nil {
		return domain.Subscription{}, err
	}

	return service.money.UpdateSubscription(ctx, domain.Subscription{
		SubscriptionID: command.SubscriptionID,
		ClientUserID:   command.ClientUserID,
		TierID:         command.TierID,
	})
}

func (service *Service) CancelSubscription(ctx context.Context, command CancelSubscriptionCommand) error {
	if err := validateSubscriptionIDCommand(command.ClientUserID, command.SubscriptionID); err != nil {
		return err
	}

	return service.money.CancelSubscription(ctx, command.ClientUserID, command.SubscriptionID)
}
