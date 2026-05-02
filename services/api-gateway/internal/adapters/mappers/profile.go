package mappers

import (
	"fmt"
	"strings"
	"time"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const publicDateLayout = "2006-01-02"

func CreateProfileRequestToProfile(
	userID int64,
	username string,
	firstName string,
	lastName string,
	isTrainer bool,
	trainerDetails *gatewayv1.TrainerDetails,
) (*profilev1.CreateProfileRequest, error) {
	mappedTrainerDetails, err := trainerDetailsToProfile(trainerDetails)
	if err != nil {
		return nil, err
	}

	return &profilev1.CreateProfileRequest{
		UserId:         userID,
		Username:       username,
		FirstName:      firstName,
		LastName:       lastName,
		IsTrainer:      isTrainer,
		TrainerDetails: mappedTrainerDetails,
	}, nil
}

func UpdateMyProfileRequestToProfile(
	userID int64,
	request *gatewayv1.UpdateMyProfileRequest,
) (*profilev1.UpdateProfileRequest, error) {
	mappedTrainerDetails, err := trainerDetailsToProfile(request.GetTrainerDetails())
	if err != nil {
		return nil, err
	}

	return &profilev1.UpdateProfileRequest{
		UserId:         userID,
		Username:       request.Username,
		FirstName:      request.FirstName,
		LastName:       request.LastName,
		Bio:            request.Bio,
		TrainerDetails: mappedTrainerDetails,
	}, nil
}

func UploadMyAvatarRequestToProfile(
	userID int64,
	request *gatewayv1.UploadMyAvatarRequest,
) *profilev1.UploadAvatarRequest {
	fileName := strings.TrimSpace(request.GetFileName())
	if fileName == "" {
		fileName = "avatar.bin"
	}

	contentType := strings.TrimSpace(request.GetContentType())
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &profilev1.UploadAvatarRequest{
		UserId:      userID,
		FileName:    fileName,
		ContentType: contentType,
		Content:     request.GetAvatar(),
	}
}

func ProfileResponseFromProfile(profile *profilev1.Profile, currentUserID int64) (*gatewayv1.ProfileResponse, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile is required")
	}

	userID, err := int64ToInt32("profile.user_id", profile.GetUserId())
	if err != nil {
		return nil, err
	}

	trainerDetails, err := trainerDetailsFromProfile(profile.GetTrainerDetails())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.ProfileResponse{
		UserId:         userID,
		IsMe:           currentUserID > 0 && currentUserID == profile.GetUserId(),
		IsTrainer:      profile.GetIsTrainer(),
		Username:       profile.GetUsername(),
		FirstName:      profile.GetFirstName(),
		LastName:       profile.GetLastName(),
		Bio:            profile.Bio,
		AvatarUrl:      profile.AvatarUrl,
		TrainerDetails: trainerDetails,
	}, nil
}

func GetTrainersResponseFromProfile(response *profilev1.SearchAuthorsResponse) (*gatewayv1.GetTrainersResponse, error) {
	if response == nil {
		return &gatewayv1.GetTrainersResponse{}, nil
	}

	trainers := make([]*gatewayv1.TrainerListItem, 0, len(response.GetAuthors()))
	for _, author := range response.GetAuthors() {
		userID, err := int64ToInt32("profile.author.user_id", author.GetUserId())
		if err != nil {
			return nil, err
		}

		trainerDetails, err := trainerDetailsFromProfile(author.GetTrainerDetails())
		if err != nil {
			return nil, err
		}

		trainers = append(trainers, &gatewayv1.TrainerListItem{
			UserId:         userID,
			IsTrainer:      true,
			Username:       author.GetUsername(),
			FirstName:      author.GetFirstName(),
			LastName:       author.GetLastName(),
			Bio:            author.Bio,
			AvatarUrl:      author.AvatarUrl,
			TrainerDetails: trainerDetails,
		})
	}

	return &gatewayv1.GetTrainersResponse{Trainers: trainers}, nil
}

func ListTrainersRequestToProfile(request *gatewayv1.ListTrainersRequest) *profilev1.SearchAuthorsRequest {
	return &profilev1.SearchAuthorsRequest{
		Query:        request.GetQuery(),
		SportTypeIds: int32SliceToInt64Slice(request.GetSportTypeIds()),
		Limit:        request.GetLimit(),
		Offset:       request.GetOffset(),
	}
}

func SportTypesResponseFromProfile(response *profilev1.ListSportTypesResponse) (*gatewayv1.SportTypesResponse, error) {
	if response == nil {
		return &gatewayv1.SportTypesResponse{}, nil
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

	return &gatewayv1.SportTypesResponse{SportTypes: sportTypes}, nil
}

func AvatarUploadResponseFromProfile(profile *profilev1.Profile) *gatewayv1.AvatarUploadResponse {
	response := &gatewayv1.AvatarUploadResponse{}
	if profile != nil && profile.AvatarUrl != nil {
		response.AvatarUrl = *profile.AvatarUrl
	}

	return response
}

func trainerDetailsToProfile(details *gatewayv1.TrainerDetails) (*profilev1.TrainerDetails, error) {
	if details == nil {
		return nil, nil
	}

	sports := make([]*profilev1.TrainerSport, 0, len(details.GetSports()))
	for _, sport := range details.GetSports() {
		sports = append(sports, &profilev1.TrainerSport{
			SportTypeId:     int32ToInt64(sport.GetSportTypeId()),
			ExperienceYears: sport.GetExperienceYears(),
			SportsRank:      sport.SportsRank,
		})
	}

	response := &profilev1.TrainerDetails{
		EducationDegree: details.EducationDegree,
		Sports:          sports,
	}
	if details.CareerSinceDate != nil {
		parsedDate, err := time.Parse(publicDateLayout, details.GetCareerSinceDate())
		if err != nil {
			return nil, fmt.Errorf("invalid career_since_date: %w", err)
		}
		response.CareerSinceDate = timestamppb.New(parsedDate.UTC())
	}

	return response, nil
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

	response := &gatewayv1.TrainerDetails{
		EducationDegree: details.EducationDegree,
		Sports:          sports,
	}
	if details.CareerSinceDate != nil {
		formattedDate := details.GetCareerSinceDate().AsTime().UTC().Format(publicDateLayout)
		response.CareerSinceDate = &formattedDate
	}

	return response, nil
}
