package postgres

import "testing"

func TestEscapeLikePattern(t *testing.T) {
	t.Parallel()

	input := `50%_coach\name`
	want := `50\%\_coach\\name`

	if got := escapeLikePattern(input); got != want {
		t.Fatalf("escapeLikePattern() = %q, want %q", got, want)
	}
}
