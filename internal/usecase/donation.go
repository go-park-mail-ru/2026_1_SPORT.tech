package usecase

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

var ErrDonationRecipientNotFound = errors.New("donation recipient not found")

type CreateDonationCommand struct {
	SenderUserID    int64
	RecipientUserID int64
	AmountValue     int64
	Currency        string
	Message         *string
}

type CreateDonationParams struct {
	SenderUserID    int64
	RecipientUserID int64
	AmountMantissa  int64
	AmountScale     int
	Currency        string
	Message         *string
}

type DonationUseCase struct {
	donationRepository donationRepository
}

func NewDonationUseCase(donationRepository donationRepository) *DonationUseCase {
	return &DonationUseCase{
		donationRepository: donationRepository,
	}
}

func (useCase *DonationUseCase) Create(ctx context.Context, command CreateDonationCommand) (domain.Donation, error) {
	donation, err := useCase.donationRepository.Create(ctx, CreateDonationParams{
		SenderUserID:    command.SenderUserID,
		RecipientUserID: command.RecipientUserID,
		AmountMantissa:  command.AmountValue,
		AmountScale:     donationCurrencyScale(command.Currency),
		Currency:        command.Currency,
		Message:         command.Message,
	})
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domain.Donation{}, ErrDonationRecipientNotFound
		}

		return domain.Donation{}, err
	}

	return donation, nil
}

func donationCurrencyScale(currency string) int {
	switch currency {
	case "JPY":
		return 0
	default:
		return 2
	}
}
