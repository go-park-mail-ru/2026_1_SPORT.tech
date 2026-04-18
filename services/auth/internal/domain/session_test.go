package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
)

func TestSessionIsActive(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)

	active := domain.Session{ExpiresAt: now.Add(time.Hour)}
	if !active.IsActive(now) {
		t.Fatal("expected active session")
	}

	expired := domain.Session{ExpiresAt: now}
	if expired.IsActive(now) {
		t.Fatal("expected expired session to be inactive")
	}

	revokedAt := now.Add(-time.Minute)
	revoked := domain.Session{
		ExpiresAt: now.Add(time.Hour),
		RevokedAt: &revokedAt,
	}
	if revoked.IsActive(now) {
		t.Fatal("expected revoked session to be inactive")
	}
}

func TestAccountCanAuthenticate(t *testing.T) {
	account := domain.Account{Status: domain.StatusActive}
	if err := account.CanAuthenticate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	disabled := domain.Account{Status: domain.StatusDisabled}
	if err := disabled.CanAuthenticate(); !errors.Is(err, domain.ErrAccountDisabled) {
		t.Fatalf("unexpected error: got %v, want %v", err, domain.ErrAccountDisabled)
	}
}
