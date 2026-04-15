package requestctx

import (
	"context"
	"log/slog"
	"testing"
)

func TestRequestContextHelpers(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	ctx = WithRequestID(ctx, "req-123")
	ctx = WithLogger(ctx, logger)

	requestID, ok := RequestIDFromContext(ctx)
	if !ok || requestID != "req-123" {
		t.Fatalf("unexpected request id: %q %v", requestID, ok)
	}

	gotLogger, ok := LoggerFromContext(ctx)
	if !ok || gotLogger == nil {
		t.Fatal("expected logger from context")
	}
}
