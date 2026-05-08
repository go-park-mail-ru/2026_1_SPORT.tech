package usecase

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

func hashSessionToken(sessionToken string) string {
	sum := sha256.Sum256([]byte(sessionToken))
	return hex.EncodeToString(sum[:])
}
