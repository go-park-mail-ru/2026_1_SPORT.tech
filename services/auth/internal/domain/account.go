package domain

import "time"

type Role string

const (
	RoleClient  Role = "client"
	RoleTrainer Role = "trainer"
	RoleAdmin   Role = "admin"
)

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type Account struct {
	ID           int64
	Email        string
	Username     string
	PasswordHash string
	Role         Role
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (role Role) IsValid() bool {
	switch role {
	case RoleClient, RoleTrainer, RoleAdmin:
		return true
	default:
		return false
	}
}

func ParseRole(value string) (Role, error) {
	role := Role(value)
	if !role.IsValid() {
		return "", ErrInvalidRole
	}

	return role, nil
}

func (status Status) IsValid() bool {
	switch status {
	case StatusActive, StatusDisabled:
		return true
	default:
		return false
	}
}

func ParseStatus(value string) (Status, error) {
	status := Status(value)
	if !status.IsValid() {
		return "", ErrInvalidStatus
	}

	return status, nil
}

func (account Account) CanAuthenticate() error {
	if account.Status != StatusActive {
		return ErrAccountDisabled
	}

	return nil
}
