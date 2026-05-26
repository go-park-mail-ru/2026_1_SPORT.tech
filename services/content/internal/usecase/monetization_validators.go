package usecase

import "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"

func validateSubscriptionTier(tier domain.SubscriptionTier) error {
	if tier.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if tier.TierID < 0 {
		return ErrInvalidSubscriptionTierID
	}
	if len(tier.Name) == 0 || len(tier.Name) > maxTierNameLen {
		return ErrInvalidSubscriptionTierName
	}
	if tier.Price < 0 {
		return ErrInvalidSubscriptionTierPrice
	}
	if tier.Description != nil && len(*tier.Description) > maxTierDescLen {
		return ErrInvalidSubscriptionTierDescription
	}

	return nil
}

func validateUpdateSubscriptionTierCommand(command UpdateSubscriptionTierCommand) error {
	if command.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.TierID <= 0 {
		return ErrInvalidSubscriptionTierID
	}
	if command.Description != nil && command.ClearDescription {
		return ErrConflictingTierDescriptionUpdate
	}

	return nil
}

func validateSubscriptionTierIDCommand(trainerUserID int64, tierID int64) error {
	if trainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if tierID <= 0 {
		return ErrInvalidSubscriptionTierID
	}

	return nil
}

func validateSubscribeToTrainerCommand(command SubscribeToTrainerCommand) error {
	if command.ClientUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.ClientUserID == command.TrainerUserID {
		return ErrInvalidSubscriptionTarget
	}
	if command.TierID <= 0 {
		return ErrInvalidSubscriptionTierID
	}

	return nil
}

func validateUpdateSubscriptionCommand(command UpdateSubscriptionCommand) error {
	if command.ClientUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.SubscriptionID <= 0 {
		return ErrInvalidSubscriptionID
	}
	if command.TierID <= 0 {
		return ErrInvalidSubscriptionTierID
	}

	return nil
}

func validateSubscriptionIDCommand(clientUserID int64, subscriptionID int64) error {
	if clientUserID <= 0 {
		return ErrInvalidUserID
	}
	if subscriptionID <= 0 {
		return ErrInvalidSubscriptionID
	}

	return nil
}

func validateDonateToProfileCommand(command DonateToProfileCommand) error {
	if command.SenderUserID <= 0 || command.RecipientUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.SenderUserID == command.RecipientUserID {
		return ErrInvalidDonationTarget
	}
	if command.AmountValue < minDonationAmount || command.AmountValue > maxDonationAmount {
		return ErrInvalidDonationAmount
	}
	if normalizeCurrency(command.Currency) != defaultCurrency {
		return ErrInvalidDonationCurrency
	}
	if command.Message != nil && len(normalizeRequiredText(*command.Message)) > maxDonationMessageLen {
		return ErrInvalidDonationMessage
	}

	return nil
}

func validateGetBalanceQuery(query GetBalanceQuery) error {
	if query.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if normalizeCurrency(query.Currency) != defaultCurrency {
		return ErrInvalidDonationCurrency
	}

	return nil
}

func validateGetTrainerStatisticsQuery(query GetTrainerStatisticsQuery) error {
	if query.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if normalizeCurrency(query.Currency) != defaultCurrency {
		return ErrInvalidDonationCurrency
	}

	return nil
}
