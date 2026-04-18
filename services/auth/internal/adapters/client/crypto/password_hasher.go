package crypto

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct {
	cost int
}

func NewPasswordHasher(cost int) *PasswordHasher {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}

	return &PasswordHasher{cost: cost}
}

func (hasher *PasswordHasher) Hash(ctx context.Context, password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), hasher.cost)
	if err != nil {
		return "", err
	}

	return string(passwordHash), nil
}

func (hasher *PasswordHasher) Compare(ctx context.Context, passwordHash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}
