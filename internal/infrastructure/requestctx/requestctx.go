package requestctx

import (
	"context"
	"log/slog"
)

type requestIDKey struct{}
type loggerKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey{}).(string)
	return requestID, ok
}

func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

func LoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	logger, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	return logger, ok
}
