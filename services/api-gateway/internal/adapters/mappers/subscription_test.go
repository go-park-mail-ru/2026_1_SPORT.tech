package mappers

import (
	"testing"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
)

func TestSubscriberFromContent(t *testing.T) {
	sub, err := SubscriberFromContent(&contentv1.Subscription{
		SubscriptionId: 1,
		ClientUserId:   10,
		TrainerUserId:  20,
		TierId:         3,
		TierName:       "Gold",
		Price:          999,
		Active:         true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.GetSubscriptionId() != 1 || sub.GetClientId() != 10 || sub.GetTierId() != 3 {
		t.Fatalf("unexpected subscriber: %+v", sub)
	}
	if sub.GetTierName() != "Gold" || sub.GetPrice() != 999 || !sub.GetActive() {
		t.Fatalf("unexpected subscriber fields: %+v", sub)
	}
}

func TestSubscriberFromContentNil(t *testing.T) {
	_, err := SubscriberFromContent(nil)
	if err == nil {
		t.Fatal("expected error for nil subscription")
	}
}

func TestSubscribersResponseFromContent(t *testing.T) {
	resp, err := SubscribersResponseFromContent(&contentv1.ListTrainerSubscribersResponse{
		Subscribers: []*contentv1.Subscription{
			{SubscriptionId: 1, ClientUserId: 10, TrainerUserId: 20, TierId: 3, TierName: "Gold"},
			{SubscriptionId: 2, ClientUserId: 11, TrainerUserId: 20, TierId: 1, TierName: "Basic"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetSubscribers()) != 2 {
		t.Fatalf("unexpected subscribers count: %d", len(resp.GetSubscribers()))
	}
}

func TestSubscribersResponseFromContentNil(t *testing.T) {
	resp, err := SubscribersResponseFromContent(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetSubscribers()) != 0 {
		t.Fatalf("expected empty subscribers, got: %+v", resp.GetSubscribers())
	}
}

func TestSubscribersResponseFromContentEmpty(t *testing.T) {
	resp, err := SubscribersResponseFromContent(&contentv1.ListTrainerSubscribersResponse{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.GetSubscribers()) != 0 {
		t.Fatalf("expected empty subscribers: %+v", resp.GetSubscribers())
	}
}
