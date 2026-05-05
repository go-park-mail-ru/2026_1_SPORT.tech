package usecase

import (
	"regexp"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,30}$`)

func validateUserID(userID int64) error {
	if userID <= 0 {
		return ErrInvalidUserID
	}

	return nil
}

func validateSearchAuthorsQuery(query SearchAuthorsQuery) error {
	if query.Limit < 0 || query.Limit > 100 {
		return ErrInvalidSearchLimit
	}
	if query.Offset < 0 {
		return ErrInvalidSearchOffset
	}
	if query.MinExperienceYears != nil && *query.MinExperienceYears < 0 {
		return ErrInvalidExperienceYears
	}
	if query.MaxExperienceYears != nil && *query.MaxExperienceYears < 0 {
		return ErrInvalidExperienceYears
	}
	if query.MinExperienceYears != nil && query.MaxExperienceYears != nil &&
		*query.MinExperienceYears > *query.MaxExperienceYears {
		return ErrInvalidExperienceYears
	}

	return nil
}

func validateUploadAvatarCommand(command UploadAvatarCommand) error {
	if err := validateUserID(command.UserID); err != nil {
		return err
	}
	if strings.TrimSpace(command.FileName) == "" {
		return ErrAvatarFileNameRequired
	}
	if strings.TrimSpace(command.ContentType) == "" {
		return ErrAvatarContentTypeRequired
	}
	if len(command.Content) == 0 {
		return ErrAvatarContentRequired
	}

	return nil
}

func validateProfile(profile domain.Profile) error {
	if err := validateUserID(profile.UserID); err != nil {
		return err
	}
	if !usernamePattern.MatchString(profile.Username) {
		return ErrInvalidUsername
	}
	if len(profile.FirstName) == 0 || len(profile.FirstName) > 100 {
		return ErrInvalidFirstName
	}
	if len(profile.LastName) == 0 || len(profile.LastName) > 100 {
		return ErrInvalidLastName
	}
	if profile.Bio != nil && len(*profile.Bio) > 1000 {
		return ErrInvalidBio
	}
	if !profile.IsTrainer && profile.TrainerDetails != nil {
		return domain.ErrTrainerProfileForbidden
	}
	if profile.TrainerDetails != nil {
		if profile.TrainerDetails.EducationDegree != nil && len(*profile.TrainerDetails.EducationDegree) > 255 {
			return ErrInvalidEducationDegree
		}
		if profile.TrainerDetails.CareerSinceDate != nil && profile.TrainerDetails.CareerSinceDate.After(time.Now().UTC()) {
			return ErrInvalidCareerSinceDate
		}
		for _, sport := range profile.TrainerDetails.Sports {
			if sport.SportTypeID <= 0 || sport.ExperienceYears < 0 {
				return ErrInvalidExperienceYears
			}
		}
	}

	return nil
}
