package mappers

import (
	"testing"
	"time"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestValidateRegisterRequests(t *testing.T) {
	if err := ValidateClientRegisterRequest(&gatewayv1.ClientRegisterRequest{
		Username:       "client_1",
		Password:       "password123",
		PasswordRepeat: "password123",
		FirstName:      "Анна",
		LastName:       "Иванова",
	}); err != nil {
		t.Fatalf("unexpected client validation error: %v", err)
	}

	if err := ValidateClientRegisterRequest(&gatewayv1.ClientRegisterRequest{
		Username:       "client_1",
		Password:       "password123",
		PasswordRepeat: "different",
		FirstName:      "Анна",
		LastName:       "Иванова",
	}); err == nil {
		t.Fatal("expected password mismatch error")
	}

	if err := ValidateClientRegisterRequest(&gatewayv1.ClientRegisterRequest{
		Username:       "bad name",
		Password:       "password123",
		PasswordRepeat: "password123",
		FirstName:      "Анна",
		LastName:       "Иванова",
	}); err == nil {
		t.Fatal("expected invalid username error")
	}

	careerSince := "2020-01-02"
	if err := ValidateTrainerRegisterRequest(&gatewayv1.TrainerRegisterRequest{
		Username:       "trainer_1",
		Password:       "password123",
		PasswordRepeat: "password123",
		FirstName:      "Петр",
		LastName:       "Смирнов",
		TrainerDetails: &gatewayv1.TrainerDetails{
			CareerSinceDate: &careerSince,
			Sports: []*gatewayv1.TrainerSport{
				{SportTypeId: 1, ExperienceYears: 4, SportsRank: stringPtr("КМС")},
			},
		},
	}); err != nil {
		t.Fatalf("unexpected trainer validation error: %v", err)
	}

	if err := ValidateTrainerRegisterRequest(&gatewayv1.TrainerRegisterRequest{
		Username:       "trainer_1",
		Password:       "password123",
		PasswordRepeat: "password123",
		FirstName:      "Петр",
		LastName:       "Смирнов",
		TrainerDetails: &gatewayv1.TrainerDetails{
			Sports: []*gatewayv1.TrainerSport{{SportTypeId: 0, ExperienceYears: 1}},
		},
	}); err == nil {
		t.Fatal("expected invalid sport error")
	}
}

func TestProfileRequestMappers(t *testing.T) {
	careerSince := "2020-01-02"
	create, err := CreateProfileRequestToProfile(
		1001,
		"trainer_1",
		"Петр",
		"Смирнов",
		true,
		&gatewayv1.TrainerDetails{
			EducationDegree: stringPtr("РГУФК"),
			CareerSinceDate: &careerSince,
			Sports: []*gatewayv1.TrainerSport{
				{SportTypeId: 2, ExperienceYears: 6, SportsRank: stringPtr("МС")},
			},
		},
	)
	if err != nil {
		t.Fatalf("unexpected create mapper error: %v", err)
	}
	if create.GetUserId() != 1001 || create.GetUsername() != "trainer_1" || !create.GetIsTrainer() {
		t.Fatalf("unexpected create profile request: %+v", create)
	}
	if create.GetTrainerDetails().GetCareerSinceDate().AsTime().UTC().Format(publicDateLayout) != careerSince ||
		create.GetTrainerDetails().GetSports()[0].GetSportTypeId() != 2 {
		t.Fatalf("unexpected trainer details: %+v", create.GetTrainerDetails())
	}

	username := "new_name"
	firstName := "Иван"
	lastName := "Петров"
	bio := "Тренер по бегу"
	update, err := UpdateMyProfileRequestToProfile(1001, &gatewayv1.UpdateMyProfileRequest{
		Username:  &username,
		FirstName: &firstName,
		LastName:  &lastName,
		Bio:       &bio,
	})
	if err != nil {
		t.Fatalf("unexpected update mapper error: %v", err)
	}
	if update.GetUserId() != 1001 || update.GetUsername() != "new_name" || update.GetBio() != "Тренер по бегу" {
		t.Fatalf("unexpected update profile request: %+v", update)
	}

	avatar := UploadMyAvatarRequestToProfile(1001, &gatewayv1.UploadMyAvatarRequest{
		FileName:    " avatar.jpg ",
		ContentType: " image/jpeg ",
		Avatar:      []byte("file"),
	})
	if avatar.GetUserId() != 1001 || avatar.GetFileName() != "avatar.jpg" || avatar.GetContentType() != "image/jpeg" {
		t.Fatalf("unexpected avatar upload request: %+v", avatar)
	}
	defaultAvatar := UploadMyAvatarRequestToProfile(1001, &gatewayv1.UploadMyAvatarRequest{})
	if defaultAvatar.GetFileName() != "avatar.bin" || defaultAvatar.GetContentType() != "application/octet-stream" {
		t.Fatalf("unexpected default avatar request: %+v", defaultAvatar)
	}
}

func TestProfileResponseMappers(t *testing.T) {
	careerSince := timestamppb.New(time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC))
	profile := &profilev1.Profile{
		UserId:    1001,
		Username:  "trainer_1",
		FirstName: "Петр",
		LastName:  "Смирнов",
		Bio:       stringPtr("bio"),
		AvatarUrl: stringPtr("https://cdn/avatar.jpg"),
		IsTrainer: true,
		TrainerDetails: &profilev1.TrainerDetails{
			EducationDegree: stringPtr("РГУФК"),
			CareerSinceDate: careerSince,
			Sports: []*profilev1.TrainerSport{
				{SportTypeId: 2, ExperienceYears: 6, SportsRank: stringPtr("МС")},
			},
		},
	}

	response, err := ProfileResponseFromProfile(profile, 1001)
	if err != nil {
		t.Fatalf("unexpected profile response error: %v", err)
	}
	if response.GetUserId() != 1001 || !response.GetIsMe() || response.GetTrainerDetails().GetCareerSinceDate() != "2020-01-02" {
		t.Fatalf("unexpected profile response: %+v", response)
	}
	if _, err := ProfileResponseFromProfile(nil, 0); err == nil {
		t.Fatal("expected nil profile error")
	}

	trainers, err := GetTrainersResponseFromProfile(&profilev1.SearchAuthorsResponse{
		Authors: []*profilev1.AuthorSummary{{
			UserId:         1002,
			Username:       "coach",
			FirstName:      "Мария",
			LastName:       "Орлова",
			TrainerDetails: profile.GetTrainerDetails(),
		}},
	})
	if err != nil {
		t.Fatalf("unexpected trainers response error: %v", err)
	}
	if len(trainers.GetTrainers()) != 1 || trainers.GetTrainers()[0].GetUserId() != 1002 {
		t.Fatalf("unexpected trainers response: %+v", trainers)
	}

	sports, err := SportTypesResponseFromProfile(&profilev1.ListSportTypesResponse{
		SportTypes: []*profilev1.SportType{{SportTypeId: 1, Name: "Бег"}},
	})
	if err != nil {
		t.Fatalf("unexpected sport types response error: %v", err)
	}
	if len(sports.GetSportTypes()) != 1 || sports.GetSportTypes()[0].GetName() != "Бег" {
		t.Fatalf("unexpected sport types response: %+v", sports)
	}

	avatar := AvatarUploadResponseFromProfile(profile)
	if avatar.GetAvatarUrl() != "https://cdn/avatar.jpg" {
		t.Fatalf("unexpected avatar response: %+v", avatar)
	}
	if AvatarUploadResponseFromProfile(nil).GetAvatarUrl() != "" {
		t.Fatal("expected empty avatar response")
	}
}

func TestListTrainersRequestToProfile(t *testing.T) {
	minExperienceYears := int32(5)
	maxExperienceYears := int32(10)
	request := &gatewayv1.ListTrainersRequest{
		Query:              "anna",
		SportTypeIds:       []int32{1, 2},
		MinExperienceYears: &minExperienceYears,
		MaxExperienceYears: &maxExperienceYears,
		OnlyWithRank:       true,
		Limit:              10,
		Offset:             20,
	}

	mapped := ListTrainersRequestToProfile(request)

	if mapped.GetQuery() != "anna" || mapped.GetLimit() != 10 || mapped.GetOffset() != 20 {
		t.Fatalf("unexpected search request: %+v", mapped)
	}
	if len(mapped.GetSportTypeIds()) != 2 || mapped.GetSportTypeIds()[0] != 1 || mapped.GetSportTypeIds()[1] != 2 {
		t.Fatalf("unexpected sport filters: %+v", mapped.GetSportTypeIds())
	}
	if mapped.MinExperienceYears == nil || *mapped.MinExperienceYears != 5 ||
		mapped.MaxExperienceYears == nil || *mapped.MaxExperienceYears != 10 ||
		!mapped.GetOnlyWithRank() {
		t.Fatalf("unexpected trainer filters: %+v", mapped)
	}
}
