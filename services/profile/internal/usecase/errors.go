package usecase

import "errors"

var (
	ErrInvalidUserID             = errors.New("invalid user id")
	ErrInvalidUsername           = errors.New("invalid username")
	ErrInvalidFirstName          = errors.New("invalid first name")
	ErrInvalidLastName           = errors.New("invalid last name")
	ErrInvalidBio                = errors.New("invalid bio")
	ErrInvalidEducationDegree    = errors.New("invalid education degree")
	ErrInvalidCareerSinceDate    = errors.New("invalid career since date")
	ErrInvalidExperienceYears    = errors.New("invalid experience years")
	ErrInvalidSearchLimit        = errors.New("invalid search limit")
	ErrInvalidSearchOffset       = errors.New("invalid search offset")
	ErrAvatarFileNameRequired    = errors.New("avatar file name is required")
	ErrAvatarContentTypeRequired = errors.New("avatar content type is required")
	ErrAvatarContentRequired     = errors.New("avatar content is required")
	ErrAvatarStorageUnavailable  = errors.New("avatar storage is unavailable")

	ErrMeasurementNotFound    = errors.New("measurement not found")
	ErrInvalidMeasuredAt      = errors.New("invalid measured_at date")
	ErrInvalidMeasurementData = errors.New("at least one measurement value must be provided")
)
