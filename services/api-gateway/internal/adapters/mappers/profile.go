package mappers

import (
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
)

func CreateProfileRequestToProfile(request *gatewayv1.CreateProfileRequest) *profilev1.CreateProfileRequest {
	return &profilev1.CreateProfileRequest{
		UserId:         request.GetUserId(),
		Username:       request.GetUsername(),
		FirstName:      request.GetFirstName(),
		LastName:       request.GetLastName(),
		Bio:            request.Bio,
		IsTrainer:      request.GetIsTrainer(),
		TrainerDetails: trainerDetailsToProfile(request.GetTrainerDetails()),
	}
}

func GetProfileRequestToProfile(request *gatewayv1.GetProfileRequest) *profilev1.GetProfileRequest {
	return &profilev1.GetProfileRequest{UserId: request.GetUserId()}
}

func UpdateProfileRequestToProfile(request *gatewayv1.UpdateProfileRequest) *profilev1.UpdateProfileRequest {
	return &profilev1.UpdateProfileRequest{
		UserId:         request.GetUserId(),
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
		SportTypeIds: request.GetSportTypeIds(),
		Limit:        request.GetLimit(),
		Offset:       request.GetOffset(),
	}
}

func UploadAvatarRequestToProfile(request *gatewayv1.UploadAvatarRequest) *profilev1.UploadAvatarRequest {
	return &profilev1.UploadAvatarRequest{
		UserId:      request.GetUserId(),
		FileName:    request.GetFileName(),
		ContentType: request.GetContentType(),
		Content:     request.GetContent(),
	}
}

func DeleteAvatarRequestToProfile(request *gatewayv1.DeleteAvatarRequest) *profilev1.DeleteAvatarRequest {
	return &profilev1.DeleteAvatarRequest{UserId: request.GetUserId()}
}

func ProfileResponseFromProfile(response *profilev1.ProfileResponse) *gatewayv1.ProfileResponse {
	if response == nil {
		return nil
	}

	return &gatewayv1.ProfileResponse{Profile: profileFromProfile(response.GetProfile())}
}

func SearchAuthorsResponseFromProfile(response *profilev1.SearchAuthorsResponse) *gatewayv1.SearchAuthorsResponse {
	if response == nil {
		return nil
	}

	authors := make([]*gatewayv1.AuthorSummary, 0, len(response.GetAuthors()))
	for _, author := range response.GetAuthors() {
		authors = append(authors, authorSummaryFromProfile(author))
	}

	return &gatewayv1.SearchAuthorsResponse{Authors: authors}
}

func ListSportTypesResponseFromProfile(response *profilev1.ListSportTypesResponse) *gatewayv1.ListSportTypesResponse {
	if response == nil {
		return nil
	}

	sportTypes := make([]*gatewayv1.SportType, 0, len(response.GetSportTypes()))
	for _, sportType := range response.GetSportTypes() {
		sportTypes = append(sportTypes, &gatewayv1.SportType{
			SportTypeId: sportType.GetSportTypeId(),
			Name:        sportType.GetName(),
		})
	}

	return &gatewayv1.ListSportTypesResponse{SportTypes: sportTypes}
}

func trainerDetailsToProfile(details *gatewayv1.TrainerDetails) *profilev1.TrainerDetails {
	if details == nil {
		return nil
	}

	sports := make([]*profilev1.TrainerSport, 0, len(details.GetSports()))
	for _, sport := range details.GetSports() {
		sports = append(sports, &profilev1.TrainerSport{
			SportTypeId:     sport.GetSportTypeId(),
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

func trainerDetailsFromProfile(details *profilev1.TrainerDetails) *gatewayv1.TrainerDetails {
	if details == nil {
		return nil
	}

	sports := make([]*gatewayv1.TrainerSport, 0, len(details.GetSports()))
	for _, sport := range details.GetSports() {
		sports = append(sports, &gatewayv1.TrainerSport{
			SportTypeId:     sport.GetSportTypeId(),
			ExperienceYears: sport.GetExperienceYears(),
			SportsRank:      sport.SportsRank,
		})
	}

	return &gatewayv1.TrainerDetails{
		EducationDegree: details.EducationDegree,
		CareerSinceDate: details.CareerSinceDate,
		Sports:          sports,
	}
}

func profileFromProfile(profile *profilev1.Profile) *gatewayv1.Profile {
	if profile == nil {
		return nil
	}

	return &gatewayv1.Profile{
		UserId:         profile.GetUserId(),
		Username:       profile.GetUsername(),
		FirstName:      profile.GetFirstName(),
		LastName:       profile.GetLastName(),
		Bio:            profile.Bio,
		AvatarUrl:      profile.AvatarUrl,
		IsTrainer:      profile.GetIsTrainer(),
		CreatedAt:      profile.GetCreatedAt(),
		UpdatedAt:      profile.GetUpdatedAt(),
		TrainerDetails: trainerDetailsFromProfile(profile.GetTrainerDetails()),
	}
}

func authorSummaryFromProfile(author *profilev1.AuthorSummary) *gatewayv1.AuthorSummary {
	if author == nil {
		return nil
	}

	return &gatewayv1.AuthorSummary{
		UserId:         author.GetUserId(),
		Username:       author.GetUsername(),
		FirstName:      author.GetFirstName(),
		LastName:       author.GetLastName(),
		Bio:            author.Bio,
		AvatarUrl:      author.AvatarUrl,
		TrainerDetails: trainerDetailsFromProfile(author.GetTrainerDetails()),
	}
}
