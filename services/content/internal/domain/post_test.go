package domain_test

import (
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func TestCanViewPost(t *testing.T) {
	required := int32(2)
	viewerLevel := int32(3)

	if !domain.CanViewPost(nil, 10, 20, nil) {
		t.Fatal("public post should be visible")
	}
	if !domain.CanViewPost(&required, 10, 10, nil) {
		t.Fatal("author should see own post")
	}
	if !domain.CanViewPost(&required, 10, 20, &viewerLevel) {
		t.Fatal("viewer with enough subscription should see post")
	}
	if domain.CanViewPost(&required, 10, 20, nil) {
		t.Fatal("viewer without subscription should not see locked post")
	}
}
