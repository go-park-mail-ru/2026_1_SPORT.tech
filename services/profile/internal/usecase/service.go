package usecase

import (
	"bytes"
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
)

type Service struct {
	profileRepository   ProfileRepository
	sportTypeRepository SportTypeRepository
	avatarStorage       AvatarStorage
}

func NewService(
	profileRepository ProfileRepository,
	sportTypeRepository SportTypeRepository,
	avatarStorage AvatarStorage,
) *Service {
	return &Service{
		profileRepository:   profileRepository,
		sportTypeRepository: sportTypeRepository,
		avatarStorage:       avatarStorage,
	}
}

func (service *Service) CreateProfile(ctx context.Context, command CreateProfileCommand) (domain.Profile, error) {
	profile, err := buildProfile(command)
	if err != nil {
		return domain.Profile{}, err
	}

	if err := service.profileRepository.Create(ctx, profile); err != nil {
		return domain.Profile{}, err
	}

	return service.profileRepository.GetByID(ctx, command.UserID)
}

func (service *Service) GetProfile(ctx context.Context, userID int64) (domain.Profile, error) {
	if err := validateUserID(userID); err != nil {
		return domain.Profile{}, err
	}

	return service.profileRepository.GetByID(ctx, userID)
}

func (service *Service) UpdateProfile(ctx context.Context, command UpdateProfileCommand) (domain.Profile, error) {
	if err := validateUserID(command.UserID); err != nil {
		return domain.Profile{}, err
	}

	profile, err := service.profileRepository.GetByID(ctx, command.UserID)
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

	if err := service.profileRepository.Update(ctx, profile); err != nil {
		return domain.Profile{}, err
	}

	return service.profileRepository.GetByID(ctx, profile.UserID)
}

func (service *Service) SearchAuthors(ctx context.Context, query SearchAuthorsQuery) ([]domain.AuthorSummary, error) {
	if err := validateSearchAuthorsQuery(query); err != nil {
		return nil, err
	}
	if query.Limit == 0 {
		query.Limit = 20
	}
	query.Query = normalizeRequiredText(query.Query)

	return service.profileRepository.SearchAuthors(ctx, query)
}

func (service *Service) UploadAvatar(ctx context.Context, command UploadAvatarCommand) (domain.Profile, error) {
	if err := validateUploadAvatarCommand(command); err != nil {
		return domain.Profile{}, err
	}
	if service.avatarStorage == nil {
		return domain.Profile{}, ErrAvatarStorageUnavailable
	}

	profile, err := service.profileRepository.GetByID(ctx, command.UserID)
	if err != nil {
		return domain.Profile{}, err
	}

	avatarURL, err := service.avatarStorage.UploadAvatar(
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

	if err := service.profileRepository.UpdateAvatarURL(ctx, command.UserID, avatarURL); err != nil {
		return domain.Profile{}, err
	}
	if profile.AvatarURL != nil {
		_ = service.avatarStorage.DeleteAvatar(ctx, *profile.AvatarURL)
	}

	return service.profileRepository.GetByID(ctx, command.UserID)
}

func (service *Service) DeleteAvatar(ctx context.Context, userID int64) error {
	if err := validateUserID(userID); err != nil {
		return err
	}
	profile, err := service.profileRepository.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if profile.AvatarURL == nil {
		return nil
	}
	if service.avatarStorage == nil {
		return ErrAvatarStorageUnavailable
	}
	if err := service.avatarStorage.DeleteAvatar(ctx, *profile.AvatarURL); err != nil {
		return err
	}

	return service.profileRepository.ClearAvatarURL(ctx, userID)
}

func (service *Service) ListSportTypes(ctx context.Context) ([]domain.SportType, error) {
	return service.sportTypeRepository.ListSportTypes(ctx)
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
