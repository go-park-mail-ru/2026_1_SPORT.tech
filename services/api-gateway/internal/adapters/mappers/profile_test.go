package mappers

import (
	"testing"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func TestListTrainersRequestToProfile(t *testing.T) {
	minExperienceYears := int32(5)
	maxExperienceYears := int32(10)
	request := &gatewayv1.ListTrainersRequest{
		Query:              "anna",
		SportTypeIds:       []int32{1, 2},
		MinExperienceYears: &minExperienceYears,
		MaxExperienceYears: &maxExperienceYears,
		OnlyWithRank:       true,
		Limit:              10,
		Offset:             20,
	}

	mapped := ListTrainersRequestToProfile(request)

	if mapped.GetQuery() != "anna" || mapped.GetLimit() != 10 || mapped.GetOffset() != 20 {
		t.Fatalf("unexpected search request: %+v", mapped)
	}
	if len(mapped.GetSportTypeIds()) != 2 || mapped.GetSportTypeIds()[0] != 1 || mapped.GetSportTypeIds()[1] != 2 {
		t.Fatalf("unexpected sport filters: %+v", mapped.GetSportTypeIds())
	}
	if mapped.MinExperienceYears == nil || *mapped.MinExperienceYears != 5 ||
		mapped.MaxExperienceYears == nil || *mapped.MaxExperienceYears != 10 ||
		!mapped.GetOnlyWithRank() {
		t.Fatalf("unexpected trainer filters: %+v", mapped)
	}
}
