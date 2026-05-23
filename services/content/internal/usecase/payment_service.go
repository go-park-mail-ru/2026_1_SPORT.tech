package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func (service *Service) DonateToProfile(ctx context.Context, command DonateToProfileCommand) (domain.Donation, error) {
	command.Currency = normalizeCurrency(command.Currency)
	command.Message = normalizeOptionalText(command.Message)
	if err := validateDonateToProfileCommand(command); err != nil {
		return domain.Donation{}, err
	}

	donation, err := service.money.CreateDonation(ctx, domain.Donation{
		SenderUserID:    command.SenderUserID,
		RecipientUserID: command.RecipientUserID,
		AmountValue:     command.AmountValue,
		Currency:        command.Currency,
		Message:         command.Message,
	})
	if err != nil {
		return domain.Donation{}, err
	}

	if err := service.createNotification(ctx, domain.Notification{
		UserID:      donation.RecipientUserID,
		Type:        domain.NotificationTypeDonation,
		ActorUserID: donation.SenderUserID,
		Title:       "Новый донат",
		Body:        "Пользователь отправил вам донат",
		DonationID:  &donation.DonationID,
	}); err != nil {
		return domain.Donation{}, err
	}

	return donation, nil
}

func (service *Service) CreateDonationPayment(ctx context.Context, command CreateDonationPaymentCommand) (domain.DonationPayment, error) {
	donationCommand := DonateToProfileCommand{
		SenderUserID:    command.SenderUserID,
		RecipientUserID: command.RecipientUserID,
		AmountValue:     command.AmountValue,
		Currency:        normalizeCurrency(command.Currency),
		Message:         normalizeOptionalText(command.Message),
	}
	if err := validateDonateToProfileCommand(donationCommand); err != nil {
		return domain.DonationPayment{}, err
	}

	confirmationToken, err := randomToken("confirm")
	if err != nil {
		return domain.DonationPayment{}, err
	}
	if service.paymentProvider == nil {
		return domain.DonationPayment{}, ErrPaymentProviderUnavailable
	}

	payment := domain.DonationPayment{
		Provider:          service.paymentProvider.ProviderName(),
		Status:            domain.PaymentStatusPending,
		SenderUserID:      donationCommand.SenderUserID,
		RecipientUserID:   donationCommand.RecipientUserID,
		AmountValue:       donationCommand.AmountValue,
		Currency:          donationCommand.Currency,
		Message:           donationCommand.Message,
		ConfirmationToken: confirmationToken,
	}
	created, err := service.money.CreateDonationPayment(ctx, payment)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	providerPayment, err := service.paymentProvider.CreatePayment(ctx, PaymentProviderCreateRequest{
		AmountValue:    donationCommand.AmountValue,
		Currency:       donationCommand.Currency,
		Description:    fmt.Sprintf("Donation payment #%d", created.PaymentID),
		IdempotenceKey: paymentIdempotenceKey("donation", created),
		ReturnURL:      normalizeOptionalURL(command.ReturnURL),
		CancelURL:      normalizeOptionalURL(command.CancelURL),
	})
	if err != nil {
		return domain.DonationPayment{}, fmt.Errorf("%w: %v", ErrPaymentProviderUnavailable, err)
	}

	return service.money.UpdateDonationPaymentProvider(ctx, created.PaymentID, providerPayment.ProviderPaymentID, providerPayment.ConfirmationURL)
}

func (service *Service) CreateSubscriptionPayment(ctx context.Context, command CreateSubscriptionPaymentCommand) (domain.DonationPayment, error) {
	if command.ClientUserID <= 0 || command.TrainerUserID <= 0 {
		return domain.DonationPayment{}, ErrInvalidUserID
	}
	if command.ClientUserID == command.TrainerUserID {
		return domain.DonationPayment{}, ErrInvalidSubscriptionTarget
	}
	if command.TierID <= 0 {
		return domain.DonationPayment{}, ErrInvalidSubscriptionTierID
	}

	tier, err := service.money.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	if tier.Price == 0 {
		return service.createFreeSubscription(ctx, command, tier.TierID)
	}

	if service.paymentProvider == nil {
		return domain.DonationPayment{}, ErrPaymentProviderUnavailable
	}

	confirmationToken, err := randomToken("confirm")
	if err != nil {
		return domain.DonationPayment{}, err
	}
	tierID := tier.TierID
	payment := domain.DonationPayment{
		Provider:          service.paymentProvider.ProviderName(),
		Status:            domain.PaymentStatusPending,
		SenderUserID:      command.ClientUserID,
		RecipientUserID:   command.TrainerUserID,
		AmountValue:       tier.Price,
		Currency:          "RUB",
		ConfirmationToken: confirmationToken,
		TierID:            &tierID,
	}
	created, err := service.money.CreateDonationPayment(ctx, payment)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	providerPayment, err := service.paymentProvider.CreatePayment(ctx, PaymentProviderCreateRequest{
		AmountValue:    tier.Price,
		Currency:       payment.Currency,
		Description:    fmt.Sprintf("Subscription payment #%d", created.PaymentID),
		IdempotenceKey: paymentIdempotenceKey("subscription", created),
		ReturnURL:      normalizeOptionalURL(command.ReturnURL),
		CancelURL:      normalizeOptionalURL(command.CancelURL),
	})
	if err != nil {
		return domain.DonationPayment{}, fmt.Errorf("%w: %v", ErrPaymentProviderUnavailable, err)
	}

	return service.money.UpdateDonationPaymentProvider(ctx, created.PaymentID, providerPayment.ProviderPaymentID, providerPayment.ConfirmationURL)
}

func (service *Service) createFreeSubscription(ctx context.Context, command CreateSubscriptionPaymentCommand, tierID int64) (domain.DonationPayment, error) {
	subscription, err := service.createPaidSubscription(ctx, SubscribeToTrainerCommand{
		ClientUserID:  command.ClientUserID,
		TrainerUserID: command.TrainerUserID,
		TierID:        command.TierID,
	})
	if err != nil {
		return domain.DonationPayment{}, err
	}

	return domain.DonationPayment{
		Status:          domain.PaymentStatusConfirmed,
		SenderUserID:    command.ClientUserID,
		RecipientUserID: command.TrainerUserID,
		AmountValue:     0,
		Currency:        "RUB",
		TierID:          &tierID,
		Subscription:    &subscription,
	}, nil
}

func (service *Service) ConfirmDonationPayment(ctx context.Context, command ConfirmDonationPaymentCommand) (domain.DonationPayment, error) {
	if command.SenderUserID <= 0 {
		return domain.DonationPayment{}, ErrInvalidUserID
	}
	if command.PaymentID <= 0 {
		return domain.DonationPayment{}, ErrInvalidPaymentID
	}
	if normalizeRequiredText(command.ConfirmationToken) == "" {
		return domain.DonationPayment{}, ErrInvalidPaymentConfirmationToken
	}

	payment, err := service.money.GetDonationPayment(ctx, command.SenderUserID, command.PaymentID)
	if err != nil {
		return domain.DonationPayment{}, err
	}
	if payment.ConfirmationToken != normalizeRequiredText(command.ConfirmationToken) {
		return domain.DonationPayment{}, domain.ErrPaymentTokenMismatch
	}
	wasPending := payment.Status != domain.PaymentStatusConfirmed
	if payment.Status != domain.PaymentStatusConfirmed {
		if service.paymentProvider == nil {
			return domain.DonationPayment{}, ErrPaymentProviderUnavailable
		}
		providerCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()

		providerPayment, err := service.paymentProvider.GetPayment(providerCtx, payment.ProviderPaymentID)
		if err != nil {
			return domain.DonationPayment{}, fmt.Errorf("%w: %v", ErrPaymentProviderUnavailable, err)
		}
		if providerPayment.Status != "succeeded" {
			return domain.DonationPayment{}, ErrPaymentNotSucceeded
		}
	}

	payment, err = service.money.ConfirmDonationPayment(ctx, command.SenderUserID, command.PaymentID, normalizeRequiredText(command.ConfirmationToken))
	if err != nil {
		return domain.DonationPayment{}, err
	}
	if wasPending && payment.Donation != nil {
		if err := service.createNotification(ctx, domain.Notification{
			UserID:      payment.Donation.RecipientUserID,
			Type:        domain.NotificationTypeDonation,
			ActorUserID: payment.Donation.SenderUserID,
			Title:       "Новый донат",
			Body:        "Пользователь отправил вам донат",
			DonationID:  &payment.Donation.DonationID,
		}); err != nil {
			return domain.DonationPayment{}, err
		}
	}
	if wasPending && payment.Subscription != nil {
		if err := service.createSubscriptionNotifications(ctx, *payment.Subscription); err != nil {
			return domain.DonationPayment{}, err
		}
	}

	return payment, nil
}

func (service *Service) ListReceivedDonations(ctx context.Context, query ListReceivedDonationsQuery) ([]domain.Donation, int32, error) {
	if query.TrainerUserID <= 0 {
		return nil, 0, ErrInvalidUserID
	}
	limit := query.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	donations, err := service.money.ListReceivedDonations(ctx, query.TrainerUserID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := service.money.CountReceivedDonations(ctx, query.TrainerUserID)
	if err != nil {
		return nil, 0, err
	}
	return donations, total, nil
}

func (service *Service) GetBalance(ctx context.Context, query GetBalanceQuery) (domain.Balance, error) {
	query.Currency = normalizeCurrency(query.Currency)
	if err := validateGetBalanceQuery(query); err != nil {
		return domain.Balance{}, err
	}

	return service.money.GetBalance(ctx, query.TrainerUserID, query.Currency)
}

func (service *Service) GetTrainerStatistics(ctx context.Context, query GetTrainerStatisticsQuery) (domain.TrainerStatistics, error) {
	query.Currency = normalizeCurrency(query.Currency)
	if err := validateGetTrainerStatisticsQuery(query); err != nil {
		return domain.TrainerStatistics{}, err
	}

	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	return service.money.GetTrainerStatistics(ctx, query.TrainerUserID, query.Currency, monthStart)
}
