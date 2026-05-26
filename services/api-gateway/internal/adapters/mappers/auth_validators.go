package mappers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

var publicUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,30}$`)

const (
	maxPublicNameLen            = 100
	maxPublicEducationDegreeLen = 255
)

func ValidateClientRegisterRequest(request *gatewayv1.ClientRegisterRequest) error {
	if request == nil {
		return fmt.Errorf("request is required")
	}
	if !PasswordsMatch(request.GetPassword(), request.GetPasswordRepeat()) {
		return fmt.Errorf("passwords do not match")
	}

	return validatePublicProfileInput(
		request.GetUsername(),
		request.GetFirstName(),
		request.GetLastName(),
		false,
		nil,
	)
}

func ValidateTrainerRegisterRequest(request *gatewayv1.TrainerRegisterRequest) error {
	if request == nil {
		return fmt.Errorf("request is required")
	}
	if !PasswordsMatch(request.GetPassword(), request.GetPasswordRepeat()) {
		return fmt.Errorf("passwords do not match")
	}

	return validatePublicProfileInput(
		request.GetUsername(),
		request.GetFirstName(),
		request.GetLastName(),
		true,
		request.GetTrainerDetails(),
	)
}

func validatePublicProfileInput(
	username string,
	firstName string,
	lastName string,
	isTrainer bool,
	trainerDetails *gatewayv1.TrainerDetails,
) error {
	if !publicUsernamePattern.MatchString(strings.TrimSpace(username)) {
		return fmt.Errorf("invalid username")
	}
	if invalidRequiredName(firstName) {
		return fmt.Errorf("invalid first name")
	}
	if invalidRequiredName(lastName) {
		return fmt.Errorf("invalid last name")
	}
	if !isTrainer && trainerDetails != nil {
		return fmt.Errorf("trainer profile forbidden")
	}

	return validatePublicTrainerDetails(trainerDetails)
}

func invalidRequiredName(value string) bool {
	normalized := strings.TrimSpace(value)
	return len(normalized) == 0 || len(normalized) > maxPublicNameLen
}

func validatePublicTrainerDetails(details *gatewayv1.TrainerDetails) error {
	if details == nil {
		return nil
	}
	if details.EducationDegree != nil && len(strings.TrimSpace(details.GetEducationDegree())) > maxPublicEducationDegreeLen {
		return fmt.Errorf("invalid education degree")
	}
	if details.CareerSinceDate != nil {
		parsedDate, err := time.Parse(publicDateLayout, details.GetCareerSinceDate())
		if err != nil {
			return fmt.Errorf("invalid career_since_date: %w", err)
		}
		if parsedDate.After(time.Now().UTC()) {
			return fmt.Errorf("invalid career since date")
		}
	}
	for _, sport := range details.GetSports() {
		if sport.GetSportTypeId() <= 0 || sport.GetExperienceYears() < 0 {
			return fmt.Errorf("invalid experience years")
		}
	}

	return nil
}
