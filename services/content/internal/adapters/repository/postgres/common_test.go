package postgres

import "testing"

func TestEscapeLikePattern(t *testing.T) {
	t.Parallel()

	input := `50%_post\title`
	want := `50\%\_post\\title`

	if got := escapeLikePattern(input); got != want {
		t.Fatalf("escapeLikePattern() = %q, want %q", got, want)
	}
}
