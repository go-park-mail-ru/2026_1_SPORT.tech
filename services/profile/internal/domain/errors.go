package domain

import "errors"

var (
	ErrProfileNotFound         = errors.New("profile not found")
	ErrProfileExists           = errors.New("profile already exists")
	ErrUsernameTaken           = errors.New("username already taken")
	ErrSportTypeNotFound       = errors.New("sport type not found")
	ErrTrainerProfileForbidden = errors.New("trainer profile forbidden")
)
