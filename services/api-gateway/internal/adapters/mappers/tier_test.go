package mappers

import (
	"testing"
	"time"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTierFromContent(t *testing.T) {
	now := timestamppb.New(time.Date(2026, time.May, 3, 12, 0, 0, 0, time.UTC))

	tier, err := TierFromContent(&contentv1.SubscriptionTier{
		TierId:      2,
		Name:        "Продвинутый",
		Price:       1500,
		Description: stringPtr("Закрытые тренировки"),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tier.GetTierId() != 2 ||
		tier.GetName() != "Продвинутый" ||
		tier.GetPrice() != 1500 ||
		tier.GetDescription() != "Закрытые тренировки" {
		t.Fatalf("unexpected tier: %+v", tier)
	}
}
