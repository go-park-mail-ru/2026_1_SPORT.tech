package usecase

import (
	"context"
	"time"
)

const stalePaymentBatchLimit = 100

type SweepResult struct {
	PendingChecked           int
	Confirmed                int
	Expired                  int
	Errors                   int
	SubscriptionsDeactivated int64
}

func (service *Service) SweepPayments(ctx context.Context, pendingTTL time.Duration, subscriptionGrace time.Duration) (SweepResult, error) {
	now := time.Now().UTC()
	var result SweepResult

	stale, err := service.money.ListStalePendingPayments(ctx, now.Add(-pendingTTL), stalePaymentBatchLimit)
	if err != nil {
		return result, err
	}

	for _, payment := range stale {
		result.PendingChecked++
		if service.paymentProvider == nil || payment.ProviderPaymentID == "" {
			continue
		}

		providerPayment, err := service.paymentProvider.GetPayment(ctx, payment.ProviderPaymentID)
		if err != nil {
			result.Errors++
			continue
		}

		if providerPayment.Status == "succeeded" {
			if _, err := service.ConfirmPaymentFromProvider(ctx, payment.ProviderPaymentID); err != nil {
				result.Errors++
				continue
			}
			result.Confirmed++
			continue
		}

		if _, err := service.ExpirePaymentFromProvider(ctx, payment.ProviderPaymentID); err != nil {
			result.Errors++
			continue
		}
		result.Expired++
	}

	deactivated, err := service.money.DeactivateExpiredSubscriptions(ctx, now.Add(-subscriptionGrace))
	if err != nil {
		return result, err
	}
	result.SubscriptionsDeactivated = deactivated

	return result, nil
}
