package usecase

import (
	"bytes"
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

type Service struct {
	profiles ProfileRepository
	authors  AuthorRepository
	avatars  AvatarRepository
	sports   SportTypeRepository
	storage  AvatarStorage
}

func NewService(
	repositories Repositories,
	avatarStorage AvatarStorage,
) *Service {
	return &Service{
		profiles: repositories.Profiles,
		authors:  repositories.Authors,
		avatars:  repositories.Avatars,
		sports:   repositories.Sports,
		storage:  avatarStorage,
	}
}

func (service *Service) CreateProfile(ctx context.Context, command CreateProfileCommand) (domain.Profile, error) {
	profile, err := buildProfile(command)
	if err != nil {
		return domain.Profile{}, err
	}

	if err := service.profiles.Create(ctx, profile); err != nil {
		return domain.Profile{}, err
	}

	return service.profiles.GetByID(ctx, command.UserID)
}

func (service *Service) GetProfile(ctx context.Context, userID int64) (domain.Profile, error) {
	if err := validateUserID(userID); err != nil {
		return domain.Profile{}, err
	}

	return service.profiles.GetByID(ctx, userID)
}

func (service *Service) UpdateProfile(ctx context.Context, command UpdateProfileCommand) (domain.Profile, error) {
	if err := validateUserID(command.UserID); err != nil {
		return domain.Profile{}, err
	}

	profile, err := service.profiles.GetByID(ctx, command.UserID)
	if err != nil {
		return domain.Profile{}, err
	}

	if command.Username != nil {
		profile.Username = normalizeRequiredText(*command.Username)
	}
	if command.FirstName != nil {
		profile.FirstName = normalizeRequiredText(*command.FirstName)
	}
	if command.LastName != nil {
		profile.LastName = normalizeRequiredText(*command.LastName)
	}
	if command.HasBio {
		profile.Bio = normalizeOptionalText(command.Bio)
	}
	if command.HasTrainerDetails {
		if err := profile.EnsureTrainer(); err != nil {
			return domain.Profile{}, err
		}
		profile.TrainerDetails = normalizeTrainerDetails(command.TrainerDetails)
	}

	if err := validateProfile(profile); err != nil {
		return domain.Profile{}, err
	}

	if err := service.profiles.Update(ctx, profile); err != nil {
		return domain.Profile{}, err
	}

	return service.profiles.GetByID(ctx, profile.UserID)
}

func (service *Service) SearchAuthors(ctx context.Context, query SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
	if err := validateSearchAuthorsQuery(query); err != nil {
		return nil, err
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	query.Query = normalizeRequiredText(query.Query)

	return service.authors.SearchAuthors(ctx, query)
}

func (service *Service) UploadAvatar(ctx context.Context, command UploadAvatarCommand) (domain.Profile, error) {
	if err := validateUploadAvatarCommand(command); err != nil {
		return domain.Profile{}, err
	}
	if service.storage == nil {
		return domain.Profile{}, ErrAvatarStorageUnavailable
	}

	profile, err := service.avatars.GetByID(ctx, command.UserID)
	if err != nil {
		return domain.Profile{}, err
	}

	avatarURL, err := service.storage.UploadAvatar(
		ctx,
		command.UserID,
		command.FileName,
		command.ContentType,
		bytes.NewReader(command.Content),
		int64(len(command.Content)),
	)
	if err != nil {
		return domain.Profile{}, err
	}

	if err := service.avatars.UpdateAvatarURL(ctx, command.UserID, avatarURL); err != nil {
		return domain.Profile{}, err
	}
	if profile.AvatarURL != nil {
		_ = service.storage.DeleteAvatar(ctx, *profile.AvatarURL)
	}

	return service.avatars.GetByID(ctx, command.UserID)
}

func (service *Service) DeleteAvatar(ctx context.Context, userID int64) error {
	if err := validateUserID(userID); err != nil {
		return err
	}
	profile, err := service.avatars.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if profile.AvatarURL == nil {
		return nil
	}
	if service.storage == nil {
		return ErrAvatarStorageUnavailable
	}
	if err := service.storage.DeleteAvatar(ctx, *profile.AvatarURL); err != nil {
		return err
	}

	return service.avatars.ClearAvatarURL(ctx, userID)
}

func (service *Service) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	return service.sports.ListSportTypes(ctx)
}

func buildProfile(command CreateProfileCommand) (domain.Profile, error) {
	profile := domain.Profile{
		UserID:         command.UserID,
		Username:       normalizeRequiredText(command.Username),
		FirstName:      normalizeRequiredText(command.FirstName),
		LastName:       normalizeRequiredText(command.LastName),
		Bio:            normalizeOptionalText(command.Bio),
		IsTrainer:      command.IsTrainer,
		TrainerDetails: normalizeTrainerDetails(command.TrainerDetails),
	}

	if err := validateProfile(profile); err != nil {
		return domain.Profile{}, err
	}

	return profile, nil
}

func normalizeRequiredText(value string) string {
	return strings.TrimSpace(value)
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeTrainerDetails(details *domain.TrainerDetails) *domain.TrainerDetails {
	if details == nil {
		return nil
	}

	normalized := &domain.TrainerDetails{
		EducationDegree: normalizeOptionalText(details.EducationDegree),
		CareerSinceDate: details.CareerSinceDate,
		Sports:          make([]domain.TrainerSport, 0, len(details.Sports)),
	}
	for _, sport := range details.Sports {
		normalized.Sports = append(normalized.Sports, domain.TrainerSport{
			SportTypeID:     sport.SportTypeID,
			ExperienceYears: sport.ExperienceYears,
			SportsRank:      normalizeOptionalText(sport.SportsRank),
		})
	}

	return normalized
}
