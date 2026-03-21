package postgres

import (
	"context"
	"database/sql"
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

type dbRunner interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type loggedRow struct {
	ctx       context.Context
	logger    *slog.Logger
	operation string
	startedAt time.Time
	row       *sql.Row
}

func (row *loggedRow) Scan(dest ...any) error {
	err := row.row.Scan(dest...)
	logDBOperation(row.ctx, row.logger, row.operation, row.startedAt, err)
	return err
}

func execContext(
	ctx context.Context,
	runner dbRunner,
	logger *slog.Logger,
	operation string,
	query string,
	args ...any,
) (sql.Result, error) {
	startedAt := time.Now()

	result, err := runner.ExecContext(ctx, query, args...)
	logDBOperation(ctx, logger, operation, startedAt, err)

	return result, err
}

func queryContext(
	ctx context.Context,
	runner dbRunner,
	logger *slog.Logger,
	operation string,
	query string,
	args ...any,
) (*sql.Rows, error) {
	startedAt := time.Now()

	rows, err := runner.QueryContext(ctx, query, args...)
	logDBOperation(ctx, logger, operation, startedAt, err)

	return rows, err
}

func queryRowContext(
	ctx context.Context,
	runner dbRunner,
	logger *slog.Logger,
	operation string,
	query string,
	args ...any,
) *loggedRow {
	return &loggedRow{
		ctx:       ctx,
		logger:    logger,
		operation: operation,
		startedAt: time.Now(),
		row:       runner.QueryRowContext(ctx, query, args...),
	}
}
