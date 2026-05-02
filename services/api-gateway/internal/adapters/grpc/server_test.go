package grpc

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestSubscriptionLevelFromContext(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-subscription-level", "2"))

	level, err := subscriptionLevelFromContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if level == nil || *level != 2 {
		t.Fatalf("unexpected subscription level: %v", level)
	}
}

func TestSubscriptionLevelFromContextRejectsInvalidValue(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-subscription-level", "bad"))

	_, err := subscriptionLevelFromContext(ctx)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("unexpected status: %s", status.Code(err))
	}
}
