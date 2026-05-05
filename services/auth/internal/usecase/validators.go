package usecase

import (
	"net/mail"
	"regexp"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/domain"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,30}$`)

func validateRegisterCommand(command RegisterCommand) error {
	if !isValidEmail(command.Email) {
		return ErrInvalidEmail
	}

	if !usernamePattern.MatchString(command.Username) {
		return ErrInvalidUsername
	}

	if len(command.Password) < 8 {
		return ErrWeakPassword
	}

	if !command.Role.IsValid() {
		return domain.ErrInvalidRole
	}

	return nil
}

func validateLoginCommand(command LoginCommand) error {
	if !isValidEmail(command.Email) {
		return ErrInvalidEmail
	}

	if len(command.Password) < 8 {
		return ErrWeakPassword
	}

	return nil
}

func validateSessionToken(sessionToken string) error {
	if strings.TrimSpace(sessionToken) == "" {
		return ErrMissingSessionToken
	}

	return nil
}

func isValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}

	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}
