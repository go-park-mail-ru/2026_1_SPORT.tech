package mappers

import (
	"errors"
	"time"

	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func CreateProfileRequestToCommand(request *profilev1.CreateProfileRequest) usecase.CreateProfileCommand {
	return usecase.CreateProfileCommand{
		UserID:         request.GetUserId(),
		Username:       request.GetUsername(),
		FirstName:      request.GetFirstName(),
		LastName:       request.GetLastName(),
		Bio:            request.Bio,
		IsTrainer:      request.GetIsTrainer(),
		TrainerDetails: trainerDetailsFromProto(request.GetTrainerDetails()),
	}
}

func UpdateProfileRequestToCommand(request *profilev1.UpdateProfileRequest) usecase.UpdateProfileCommand {
	command := usecase.UpdateProfileCommand{
		UserID:    request.GetUserId(),
		Username:  request.Username,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	}
	if request.Bio != nil {
		command.HasBio = true
		command.Bio = request.Bio
	}
	if request.TrainerDetails != nil {
		command.HasTrainerDetails = true
		command.TrainerDetails = trainerDetailsFromProto(request.GetTrainerDetails())
	}

	return command
}

func SearchAuthorsRequestToQuery(request *profilev1.SearchAuthorsRequest) usecase.SearchAuthorsQuery {
	return usecase.SearchAuthorsQuery{
		Query:              request.GetQuery(),
		SportTypeIDs:       request.GetSportTypeIds(),
		MinExperienceYears: request.MinExperienceYears,
		MaxExperienceYears: request.MaxExperienceYears,
		OnlyWithRank:       request.GetOnlyWithRank(),
		Limit:              request.GetLimit(),
		Offset:             request.GetOffset(),
	}
}

func UploadAvatarRequestToCommand(request *profilev1.UploadAvatarRequest) usecase.UploadAvatarCommand {
	return usecase.UploadAvatarCommand{
		UserID:      request.GetUserId(),
		FileName:    request.GetFileName(),
		ContentType: request.GetContentType(),
		Content:     request.GetContent(),
	}
}

func NewProfileResponse(profile domain.Profile) *profilev1.ProfileResponse {
	return &profilev1.ProfileResponse{
		Profile: profileToProto(profile),
	}
}

func NewSearchAuthorsResponse(authors []domain.AuthorSummary) *profilev1.SearchAuthorsResponse {
	response := &profilev1.SearchAuthorsResponse{
		Authors: make([]*profilev1.AuthorSummary, 0, len(authors)),
	}
	for _, author := range authors {
		response.Authors = append(response.Authors, authorSummaryToProto(author))
	}

	return response
}

func NewListSportTypesResponse(sportTypes []domain.SportType) *profilev1.ListSportTypesResponse {
	response := &profilev1.ListSportTypesResponse{
		SportTypes: make([]*profilev1.SportType, 0, len(sportTypes)),
	}
	for _, sportType := range sportTypes {
		response.SportTypes = append(response.SportTypes, &profilev1.SportType{
			SportTypeId: sportType.ID,
			Name:        sportType.Name,
		})
	}

	return response
}

func Empty() *emptypb.Empty {
	return &emptypb.Empty{}
}

func ErrorToStatus(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, usecase.ErrInvalidUserID),
		errors.Is(err, usecase.ErrInvalidUsername),
		errors.Is(err, usecase.ErrInvalidFirstName),
		errors.Is(err, usecase.ErrInvalidLastName),
		errors.Is(err, usecase.ErrInvalidBio),
		errors.Is(err, usecase.ErrInvalidEducationDegree),
		errors.Is(err, usecase.ErrInvalidCareerSinceDate),
		errors.Is(err, usecase.ErrInvalidExperienceYears),
		errors.Is(err, usecase.ErrInvalidSearchLimit),
		errors.Is(err, usecase.ErrInvalidSearchOffset),
		errors.Is(err, usecase.ErrAvatarFileNameRequired),
		errors.Is(err, usecase.ErrAvatarContentTypeRequired),
		errors.Is(err, usecase.ErrAvatarContentRequired),
		errors.Is(err, domain.ErrSportTypeNotFound):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrProfileNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrProfileExists), errors.Is(err, domain.ErrUsernameTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrTrainerProfileForbidden):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, usecase.ErrAvatarStorageUnavailable):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func profileToProto(profile domain.Profile) *profilev1.Profile {
	response := &profilev1.Profile{
		UserId:         profile.UserID,
		Username:       profile.Username,
		FirstName:      profile.FirstName,
		LastName:       profile.LastName,
		IsTrainer:      profile.IsTrainer,
		CreatedAt:      timestamppb.New(profile.CreatedAt),
		UpdatedAt:      timestamppb.New(profile.UpdatedAt),
		TrainerDetails: trainerDetailsToProto(profile.TrainerDetails),
	}
	if profile.Bio != nil {
		response.Bio = profile.Bio
	}
	if profile.AvatarURL != nil {
		response.AvatarUrl = profile.AvatarURL
	}

	return response
}

func authorSummaryToProto(author domain.AuthorSummary) *profilev1.AuthorSummary {
	response := &profilev1.AuthorSummary{
		UserId:         author.UserID,
		Username:       author.Username,
		FirstName:      author.FirstName,
		LastName:       author.LastName,
		TrainerDetails: trainerDetailsToProto(author.TrainerDetails),
	}
	if author.Bio != nil {
		response.Bio = author.Bio
	}
	if author.AvatarURL != nil {
		response.AvatarUrl = author.AvatarURL
	}

	return response
}

func trainerDetailsFromProto(details *profilev1.TrainerDetails) *domain.TrainerDetails {
	if details == nil {
		return nil
	}

	result := &domain.TrainerDetails{
		EducationDegree: details.EducationDegree,
		Sports:          make([]domain.TrainerSport, 0, len(details.GetSports())),
	}
	if details.CareerSinceDate != nil {
		timestamp := details.GetCareerSinceDate().AsTime()
		result.CareerSinceDate = &timestamp
	}
	for _, sport := range details.GetSports() {
		result.Sports = append(result.Sports, domain.TrainerSport{
			SportTypeID:     sport.GetSportTypeId(),
			ExperienceYears: int(sport.GetExperienceYears()),
			SportsRank:      sport.SportsRank,
		})
	}

	return result
}

func trainerDetailsToProto(details *domain.TrainerDetails) *profilev1.TrainerDetails {
	if details == nil {
		return nil
	}

	response := &profilev1.TrainerDetails{
		EducationDegree: details.EducationDegree,
		Sports:          make([]*profilev1.TrainerSport, 0, len(details.Sports)),
	}
	if details.CareerSinceDate != nil {
		response.CareerSinceDate = timestamppb.New(details.CareerSinceDate.UTC())
	}
	for _, sport := range details.Sports {
		response.Sports = append(response.Sports, &profilev1.TrainerSport{
			SportTypeId:     sport.SportTypeID,
			ExperienceYears: int32(sport.ExperienceYears),
			SportsRank:      sport.SportsRank,
		})
	}

	return response
}

func NormalizeDate(value time.Time) time.Time {
	return value.UTC()
}
