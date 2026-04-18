package mappers

import (
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
)

func CreateProfileRequestToProfile(request *gatewayv1.CreateProfileRequest) *profilev1.CreateProfileRequest {
	return &profilev1.CreateProfileRequest{
		UserId:         int32ToInt64(request.GetUserId()),
		Username:       request.GetUsername(),
		FirstName:      request.GetFirstName(),
		LastName:       request.GetLastName(),
		Bio:            request.Bio,
		IsTrainer:      request.GetIsTrainer(),
		TrainerDetails: trainerDetailsToProfile(request.GetTrainerDetails()),
	}
}

func GetProfileRequestToProfile(request *gatewayv1.GetProfileRequest) *profilev1.GetProfileRequest {
	return &profilev1.GetProfileRequest{UserId: int32ToInt64(request.GetUserId())}
}

func UpdateProfileRequestToProfile(request *gatewayv1.UpdateProfileRequest) *profilev1.UpdateProfileRequest {
	return &profilev1.UpdateProfileRequest{
		UserId:         int32ToInt64(request.GetUserId()),
		Username:       request.Username,
		FirstName:      request.FirstName,
		LastName:       request.LastName,
		Bio:            request.Bio,
		TrainerDetails: trainerDetailsToProfile(request.GetTrainerDetails()),
	}
}

func SearchAuthorsRequestToProfile(request *gatewayv1.SearchAuthorsRequest) *profilev1.SearchAuthorsRequest {
	return &profilev1.SearchAuthorsRequest{
		Query:        request.GetQuery(),
		SportTypeIds: int32SliceToInt64Slice(request.GetSportTypeIds()),
		Limit:        request.GetLimit(),
		Offset:       request.GetOffset(),
	}
}

func UploadAvatarRequestToProfile(request *gatewayv1.UploadAvatarRequest) *profilev1.UploadAvatarRequest {
	return &profilev1.UploadAvatarRequest{
		UserId:      int32ToInt64(request.GetUserId()),
		FileName:    request.GetFileName(),
		ContentType: request.GetContentType(),
		Content:     request.GetContent(),
	}
}

func DeleteAvatarRequestToProfile(request *gatewayv1.DeleteAvatarRequest) *profilev1.DeleteAvatarRequest {
	return &profilev1.DeleteAvatarRequest{UserId: int32ToInt64(request.GetUserId())}
}

func ProfileResponseFromProfile(response *profilev1.ProfileResponse) (*gatewayv1.ProfileResponse, error) {
	if response == nil {
		return nil, nil
	}

	profile, err := profileFromProfile(response.GetProfile())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.ProfileResponse{Profile: profile}, nil
}

func SearchAuthorsResponseFromProfile(response *profilev1.SearchAuthorsResponse) (*gatewayv1.SearchAuthorsResponse, error) {
	if response == nil {
		return nil, nil
	}

	authors := make([]*gatewayv1.AuthorSummary, 0, len(response.GetAuthors()))
	for _, author := range response.GetAuthors() {
		mappedAuthor, err := authorSummaryFromProfile(author)
		if err != nil {
			return nil, err
		}

		authors = append(authors, mappedAuthor)
	}

	return &gatewayv1.SearchAuthorsResponse{Authors: authors}, nil
}

func ListSportTypesResponseFromProfile(response *profilev1.ListSportTypesResponse) (*gatewayv1.ListSportTypesResponse, error) {
	if response == nil {
		return nil, nil
	}

	sportTypes := make([]*gatewayv1.SportType, 0, len(response.GetSportTypes()))
	for _, sportType := range response.GetSportTypes() {
		sportTypeID, err := int64ToInt32("profile.sport_type_id", sportType.GetSportTypeId())
		if err != nil {
			return nil, err
		}

		sportTypes = append(sportTypes, &gatewayv1.SportType{
			SportTypeId: sportTypeID,
			Name:        sportType.GetName(),
		})
	}

	return &gatewayv1.ListSportTypesResponse{SportTypes: sportTypes}, nil
}

func trainerDetailsToProfile(details *gatewayv1.TrainerDetails) *profilev1.TrainerDetails {
	if details == nil {
		return nil
	}

	sports := make([]*profilev1.TrainerSport, 0, len(details.GetSports()))
	for _, sport := range details.GetSports() {
		sports = append(sports, &profilev1.TrainerSport{
			SportTypeId:     int32ToInt64(sport.GetSportTypeId()),
			ExperienceYears: sport.GetExperienceYears(),
			SportsRank:      sport.SportsRank,
		})
	}

	return &profilev1.TrainerDetails{
		EducationDegree: details.EducationDegree,
		CareerSinceDate: details.CareerSinceDate,
		Sports:          sports,
	}
}

func trainerDetailsFromProfile(details *profilev1.TrainerDetails) (*gatewayv1.TrainerDetails, error) {
	if details == nil {
		return nil, nil
	}

	sports := make([]*gatewayv1.TrainerSport, 0, len(details.GetSports()))
	for _, sport := range details.GetSports() {
		sportTypeID, err := int64ToInt32("profile.trainer_sport.sport_type_id", sport.GetSportTypeId())
		if err != nil {
			return nil, err
		}

		sports = append(sports, &gatewayv1.TrainerSport{
			SportTypeId:     sportTypeID,
			ExperienceYears: sport.GetExperienceYears(),
			SportsRank:      sport.SportsRank,
		})
	}

	return &gatewayv1.TrainerDetails{
		EducationDegree: details.EducationDegree,
		CareerSinceDate: details.CareerSinceDate,
		Sports:          sports,
	}, nil
}

func profileFromProfile(profile *profilev1.Profile) (*gatewayv1.Profile, error) {
	if profile == nil {
		return nil, nil
	}

	userID, err := int64ToInt32("profile.user_id", profile.GetUserId())
	if err != nil {
		return nil, err
	}

	trainerDetails, err := trainerDetailsFromProfile(profile.GetTrainerDetails())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.Profile{
		UserId:         userID,
		Username:       profile.GetUsername(),
		FirstName:      profile.GetFirstName(),
		LastName:       profile.GetLastName(),
		Bio:            profile.Bio,
		AvatarUrl:      profile.AvatarUrl,
		IsTrainer:      profile.GetIsTrainer(),
		CreatedAt:      profile.GetCreatedAt(),
		UpdatedAt:      profile.GetUpdatedAt(),
		TrainerDetails: trainerDetails,
	}, nil
}

func authorSummaryFromProfile(author *profilev1.AuthorSummary) (*gatewayv1.AuthorSummary, error) {
	if author == nil {
		return nil, nil
	}

	userID, err := int64ToInt32("profile.author.user_id", author.GetUserId())
	if err != nil {
		return nil, err
	}

	trainerDetails, err := trainerDetailsFromProfile(author.GetTrainerDetails())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.AuthorSummary{
		UserId:         userID,
		Username:       author.GetUsername(),
		FirstName:      author.GetFirstName(),
		LastName:       author.GetLastName(),
		Bio:            author.Bio,
		AvatarUrl:      author.AvatarUrl,
		TrainerDetails: trainerDetails,
	}, nil
}
