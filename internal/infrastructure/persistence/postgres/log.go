package postgres

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/requestctx"
)

func loggerFromContext(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if logger, ok := requestctx.LoggerFromContext(ctx); ok && logger != nil {
		return logger
	}

	if fallback != nil {
		return fallback
	}

	return slog.Default()
}

func logDBOperation(ctx context.Context, logger *slog.Logger, operation string, startedAt time.Time, err error) {
	requestLogger := loggerFromContext(ctx, logger)

	args := []any{
		"component", "postgres",
		"operation", operation,
		"duration_ms", time.Since(startedAt).Milliseconds(),
	}

	if err != nil {
		requestLogger.Error("db operation failed", append(args, "error", err)...)
		return
	}

	requestLogger.Info("db operation completed", args...)
}
