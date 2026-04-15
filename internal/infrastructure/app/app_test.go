package app

import (
	"log/slog"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/config"
)

func TestRunInvalidSessionTTL(t *testing.T) {
	cfg := config.Config{
		Auth: config.AuthConfig{
			SessionTTL: "bad-duration",
		},
	}

	err := Run(cfg, nil, slog.Default())
	if err == nil {
		t.Fatal("expected error")
	}
}
