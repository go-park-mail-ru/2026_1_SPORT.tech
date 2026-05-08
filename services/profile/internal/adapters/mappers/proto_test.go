package mappers

import (
	"errors"
	"testing"
	"time"

	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProfileRequestMappers(t *testing.T) {
	careerSince := time.Date(2020, time.January, 2, 0, 0, 0, 0, time.UTC)
	bio := "bio"
	rank := "КМС"
	request := &profilev1.CreateProfileRequest{
		UserId:    1001,
		Username:  "trainer",
		FirstName: "Анна",
		LastName:  "Павлова",
		Bio:       &bio,
		IsTrainer: true,
		TrainerDetails: &profilev1.TrainerDetails{
			EducationDegree: stringPtr("РГУФК"),
			CareerSinceDate: timestamppb.New(careerSince),
			Sports: []*profilev1.TrainerSport{{
				SportTypeId:     3001,
				ExperienceYears: 7,
				SportsRank:      &rank,
			}},
		},
	}

	command := CreateProfileRequestToCommand(request)
	if command.UserID != 1001 ||
		command.Username != "trainer" ||
		command.Bio == nil ||
		*command.Bio != "bio" ||
		command.TrainerDetails == nil ||
		command.TrainerDetails.CareerSinceDate == nil ||
		!command.TrainerDetails.CareerSinceDate.Equal(careerSince) ||
		len(command.TrainerDetails.Sports) != 1 ||
		command.TrainerDetails.Sports[0].SportTypeID != 3001 {
		t.Fatalf("unexpected create command: %+v", command)
	}

	update := UpdateProfileRequestToCommand(&profilev1.UpdateProfileRequest{
		UserId:    1002,
		Username:  stringPtr("client"),
		FirstName: stringPtr("Иван"),
		LastName:  stringPtr("Иванов"),
		Bio:       &bio,
	})
	if update.UserID != 1002 || !update.HasBio || update.Bio == nil || *update.Bio != "bio" {
		t.Fatalf("unexpected update command: %+v", update)
	}

	search := SearchAuthorsRequestToQuery(&profilev1.SearchAuthorsRequest{
		Query:              "run",
		SportTypeIds:       []int64{3001},
		MinExperienceYears: int32Ptr(2),
		MaxExperienceYears: int32Ptr(10),
		OnlyWithRank:       true,
		Limit:              20,
		Offset:             5,
	})
	if search.Query != "run" ||
		len(search.SportTypeIDs) != 1 ||
		search.MinExperienceYears == nil ||
		*search.MinExperienceYears != 2 ||
		!search.OnlyWithRank ||
		search.Limit != 20 ||
		search.Offset != 5 {
		t.Fatalf("unexpected search query: %+v", search)
	}

	avatar := UploadAvatarRequestToCommand(&profilev1.UploadAvatarRequest{
		UserId:      1001,
		FileName:    "avatar.jpg",
		ContentType: "image/jpeg",
		Content:     []byte("data"),
	})
	if avatar.UserID != 1001 || avatar.FileName != "avatar.jpg" || string(avatar.Content) != "data" {
		t.Fatalf("unexpected avatar command: %+v", avatar)
	}
}

func TestProfileResponseMappers(t *testing.T) {
	now := time.Date(2026, time.May, 6, 12, 0, 0, 0, time.UTC)
	profile := domain.Profile{
		UserID:    1001,
		Username:  "trainer",
		FirstName: "Анна",
		LastName:  "Павлова",
		Bio:       stringPtr("bio"),
		AvatarURL: stringPtr("http://cdn/avatar.jpg"),
		IsTrainer: true,
		TrainerDetails: &domain.TrainerDetails{
			EducationDegree: stringPtr("РГУФК"),
			CareerSinceDate: &now,
			Sports: []domain.TrainerSport{{
				SportTypeID:     3001,
				ExperienceYears: 7,
				SportsRank:      stringPtr("КМС"),
			}},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	response := NewProfileResponse(profile)
	if response.GetProfile().GetUserId() != 1001 ||
		response.GetProfile().GetBio() != "bio" ||
		response.GetProfile().GetAvatarUrl() != "http://cdn/avatar.jpg" ||
		len(response.GetProfile().GetTrainerDetails().GetSports()) != 1 {
		t.Fatalf("unexpected profile response: %+v", response)
	}

	authors := NewSearchAuthorsResponse([]domain.AuthorSummary{{
		UserID:    1001,
		Username:  "trainer",
		FirstName: "Анна",
		LastName:  "Павлова",
		Bio:       stringPtr("bio"),
		AvatarURL: stringPtr("http://cdn/avatar.jpg"),
	}})
	if len(authors.GetAuthors()) != 1 || authors.GetAuthors()[0].GetUserId() != 1001 {
		t.Fatalf("unexpected authors response: %+v", authors)
	}

	sports := NewListSportTypesResponse([]domain.SportType{{ID: 3001, Name: "Бег"}})
	if len(sports.GetSportTypes()) != 1 || sports.GetSportTypes()[0].GetName() != "Бег" {
		t.Fatalf("unexpected sport types response: %+v", sports)
	}

	if Empty() == nil {
		t.Fatal("empty response is nil")
	}
	if !NormalizeDate(now).Equal(now.UTC()) {
		t.Fatal("date was not normalized to UTC")
	}
}

func TestProfileErrorToStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "nil", err: nil, code: codes.OK},
		{name: "invalid argument", err: usecase.ErrInvalidUserID, code: codes.InvalidArgument},
		{name: "not found", err: domain.ErrProfileNotFound, code: codes.NotFound},
		{name: "exists", err: domain.ErrProfileExists, code: codes.AlreadyExists},
		{name: "failed precondition", err: domain.ErrTrainerProfileForbidden, code: codes.FailedPrecondition},
		{name: "internal", err: errors.New("boom"), code: codes.Internal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ErrorToStatus(test.err)
			if test.err == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if got := status.Code(err); got != test.code {
				t.Fatalf("code = %s, want %s", got, test.code)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}
