package usecase

import (
	"context"
	"fmt"
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
		TrainerUserID:   command.TrainerUserID,
		Name:            normalizeRequiredText(command.Name),
		Price:           command.Price,
		Description:     normalizeOptionalText(command.Description),
		ChatEnabled:     command.ChatEnabled,
		CalendarEnabled: command.CalendarEnabled,
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

	currentTier, err := service.money.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
	if err != nil {
		return domain.SubscriptionTier{}, err
	}
	tier := currentTier

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
	if command.CalendarEnabled != nil {
		tier.CalendarEnabled = *command.CalendarEnabled
	}

	if err := validateSubscriptionTier(tier); err != nil {
		return domain.SubscriptionTier{}, err
	}

	var affectedSubscriptions []domain.Subscription
	if command.Price != nil && tier.Price > currentTier.Price {
		affectedSubscriptions, err = service.cancelRenewalsForTierPriceIncrease(ctx, tier)
		if err != nil {
			return domain.SubscriptionTier{}, err
		}
	}

	updated, err := service.money.UpdateSubscriptionTier(ctx, tier)
	if err != nil {
		return domain.SubscriptionTier{}, err
	}

	if len(affectedSubscriptions) > 0 {
		if err := service.notifySubscribersAboutTierPriceIncrease(ctx, affectedSubscriptions, updated); err != nil {
			return domain.SubscriptionTier{}, err
		}
	}

	return updated, nil
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
		TierName:      tier.Name,
		Price:         tier.Price,
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

func (service *Service) cancelRenewalsForTierPriceIncrease(ctx context.Context, tier domain.SubscriptionTier) ([]domain.Subscription, error) {
	subscriptions, err := service.money.ListSubscriptionsAffectedByTierPriceIncrease(ctx, tier.TrainerUserID, tier.TierID, tier.Price)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, nil
	}

	for _, subscription := range subscriptions {
		if !subscription.AutoRenew || subscription.StripeSubscriptionID == "" {
			continue
		}
		if service.paymentProvider == nil {
			return nil, ErrPaymentProviderUnavailable
		}
		if err := service.paymentProvider.CancelSubscription(ctx, subscription.StripeSubscriptionID, true); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrPaymentProviderUnavailable, err)
		}
	}

	for _, subscription := range subscriptions {
		if err := service.money.BlockSubscriptionRenewalForPriceIncrease(ctx, subscription.SubscriptionID); err != nil {
			return nil, err
		}
	}

	return subscriptions, nil
}

func (service *Service) notifySubscribersAboutTierPriceIncrease(ctx context.Context, subscriptions []domain.Subscription, tier domain.SubscriptionTier) error {
	for _, subscription := range subscriptions {
		periodEnd := subscription.ExpiresAt
		if subscription.CurrentPeriodEnd != nil {
			periodEnd = *subscription.CurrentPeriodEnd
		}
		body := fmt.Sprintf(
			"Цена тарифа «%s» повышена до %d ₽. Следующего списания не будет; доступ сохранится до %s. Чтобы продолжить, оформите подписку заново.",
			tier.Name,
			tier.Price,
			periodEnd.Format("02.01.2006"),
		)
		if err := service.createNotification(ctx, domain.Notification{
			UserID:         subscription.ClientUserID,
			Type:           domain.NotificationTypeSubscription,
			ActorUserID:    tier.TrainerUserID,
			Title:          "Цена подписки изменилась",
			Body:           body,
			SubscriptionID: &subscription.SubscriptionID,
		}); err != nil {
			return err
		}
	}

	return nil
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

	subscription, err := service.money.GetSubscription(ctx, command.ClientUserID, command.SubscriptionID)
	if err != nil {
		return err
	}

	if subscription.StripeSubscriptionID != "" {
		if service.paymentProvider == nil {
			return ErrPaymentProviderUnavailable
		}
		if err := service.paymentProvider.CancelSubscription(ctx, subscription.StripeSubscriptionID, true); err != nil {
			return fmt.Errorf("%w: %v", ErrPaymentProviderUnavailable, err)
		}

		return service.money.SetSubscriptionAutoRenew(ctx, command.ClientUserID, command.SubscriptionID, false)
	}

	return service.money.CancelSubscription(ctx, command.ClientUserID, command.SubscriptionID)
}
