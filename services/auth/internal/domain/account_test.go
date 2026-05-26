package domain

import (
	"errors"
	"testing"
)

func TestRoleIsValid(t *testing.T) {
	cases := []struct {
		role  Role
		valid bool
	}{
		{RoleClient, true},
		{RoleTrainer, true},
		{RoleAdmin, true},
		{Role("unknown"), false},
		{Role(""), false},
	}
	for _, tc := range cases {
		if got := tc.role.IsValid(); got != tc.valid {
			t.Errorf("Role(%q).IsValid() = %v, want %v", tc.role, got, tc.valid)
		}
	}
}

func TestParseRole(t *testing.T) {
	role, err := ParseRole("client")
	if err != nil || role != RoleClient {
		t.Fatalf("ParseRole(client): got %v, %v", role, err)
	}

	role, err = ParseRole("trainer")
	if err != nil || role != RoleTrainer {
		t.Fatalf("ParseRole(trainer): got %v, %v", role, err)
	}

	_, err = ParseRole("unknown")
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("ParseRole(unknown): got %v, want ErrInvalidRole", err)
	}
}

func TestStatusIsValid(t *testing.T) {
	cases := []struct {
		status Status
		valid  bool
	}{
		{StatusActive, true},
		{StatusDisabled, true},
		{Status("banned"), false},
		{Status(""), false},
	}
	for _, tc := range cases {
		if got := tc.status.IsValid(); got != tc.valid {
			t.Errorf("Status(%q).IsValid() = %v, want %v", tc.status, got, tc.valid)
		}
	}
}

func TestParseStatus(t *testing.T) {
	status, err := ParseStatus("active")
	if err != nil || status != StatusActive {
		t.Fatalf("ParseStatus(active): got %v, %v", status, err)
	}

	status, err = ParseStatus("disabled")
	if err != nil || status != StatusDisabled {
		t.Fatalf("ParseStatus(disabled): got %v, %v", status, err)
	}

	_, err = ParseStatus("banned")
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("ParseStatus(banned): got %v, want ErrInvalidStatus", err)
	}
}
