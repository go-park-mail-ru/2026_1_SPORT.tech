package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/golang/mock/gomock"
)

func TestDonationUseCaseCreate(t *testing.T) {
	t.Run("success uses default scale", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := NewMockdonationRepository(ctrl)
		message := "thanks"
		now := time.Now()

		repository.EXPECT().
			Create(gomock.Any(), CreateDonationParams{
				SenderUserID:    1,
				RecipientUserID: 2,
				AmountMantissa:  5000,
				AmountScale:     2,
				Currency:        "RUB",
				Message:         &message,
			}).
			Return(domain.Donation{
				DonationID:      10,
				SenderUserID:    1,
				RecipientUserID: 2,
				AmountValue:     5000,
				Currency:        "RUB",
				Message:         &message,
				CreatedAt:       now,
				UpdatedAt:       now,
			}, nil)

		useCase := NewDonationUseCase(repository)

		donation, err := useCase.Create(context.Background(), CreateDonationCommand{
			SenderUserID:    1,
			RecipientUserID: 2,
			AmountValue:     5000,
			Currency:        "RUB",
			Message:         &message,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if donation.DonationID != 10 {
			t.Fatalf("unexpected donation: %+v", donation)
		}
	})

	t.Run("success uses jpy scale zero", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := NewMockdonationRepository(ctrl)

		repository.EXPECT().
			Create(gomock.Any(), CreateDonationParams{
				SenderUserID:    1,
				RecipientUserID: 2,
				AmountMantissa:  5000,
				AmountScale:     0,
				Currency:        "JPY",
			}).
			Return(domain.Donation{DonationID: 11}, nil)

		useCase := NewDonationUseCase(repository)

		_, err := useCase.Create(context.Background(), CreateDonationCommand{
			SenderUserID:    1,
			RecipientUserID: 2,
			AmountValue:     5000,
			Currency:        "JPY",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("maps recipient not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := NewMockdonationRepository(ctrl)
		repository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.Donation{}, ErrUserNotFound)

		useCase := NewDonationUseCase(repository)

		_, err := useCase.Create(context.Background(), CreateDonationCommand{})
		if !errors.Is(err, ErrDonationRecipientNotFound) {
			t.Fatalf("unexpected error: got %v, expect %v", err, ErrDonationRecipientNotFound)
		}
	})
}
