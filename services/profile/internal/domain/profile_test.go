package domain_test

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

func TestProfileEnsureTrainer(t *testing.T) {
	trainer := domain.Profile{IsTrainer: true}
	if err := trainer.EnsureTrainer(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client := domain.Profile{IsTrainer: false}
	if err := client.EnsureTrainer(); !errors.Is(err, domain.ErrTrainerProfileForbidden) {
		t.Fatalf("unexpected error: got %v, want %v", err, domain.ErrTrainerProfileForbidden)
	}
}
