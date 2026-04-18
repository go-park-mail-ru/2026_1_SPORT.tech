package usecase

import "errors"

var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidUsername     = errors.New("invalid username")
	ErrWeakPassword        = errors.New("weak password")
	ErrMissingSessionToken = errors.New("missing session token")
)
