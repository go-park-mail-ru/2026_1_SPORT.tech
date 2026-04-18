package crypto

import (
	"context"
	"crypto/rand"
	"encoding/base64"
)

type TokenGenerator struct {
	size int
}

func NewTokenGenerator(size int) *TokenGenerator {
	if size <= 0 {
		size = 32
	}

	return &TokenGenerator{size: size}
}

func (generator *TokenGenerator) NewToken(ctx context.Context) (string, error) {
	buffer := make([]byte, generator.size)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
