package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

const maxAvatarSize = 5 * 1024 * 1024

var allowedAvatarContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

type profileResponse struct {
	UserID         int64                   `json:"user_id"`
	IsMe           bool                    `json:"is_me"`
	IsTrainer      bool                    `json:"is_trainer"`
	Username       string                  `json:"username"`
	FirstName      string                  `json:"first_name"`
	LastName       string                  `json:"last_name"`
	Bio            *string                 `json:"bio"`
	AvatarURL      *string                 `json:"avatar_url"`
	TrainerDetails *trainerDetailsResponse `json:"trainer_details"`
}

type getTrainersResponse struct {
	Trainers []trainerListItemResponse `json:"trainers"`
}

type trainerListItemResponse struct {
	UserID         int64                   `json:"user_id"`
	IsTrainer      bool                    `json:"is_trainer"`
	Username       string                  `json:"username"`
	FirstName      string                  `json:"first_name"`
	LastName       string                  `json:"last_name"`
	Bio            *string                 `json:"bio"`
	AvatarURL      *string                 `json:"avatar_url"`
	TrainerDetails *trainerDetailsResponse `json:"trainer_details"`
}

type trainerSportResponse struct {
	SportTypeID     int64   `json:"sport_type_id"`
	ExperienceYears int     `json:"experience_years"`
	SportsRank      *string `json:"sports_rank"`
}

type trainerDetailsResponse struct {
	EducationDegree *string                `json:"education_degree"`
	CareerSinceDate string                 `json:"career_since_date"`
	Sports          []trainerSportResponse `json:"sports"`
}

type avatarUploadResponse struct {
	AvatarURL string `json:"avatar_url"`
}

func (handler *Handler) handleGetTrainers(writer http.ResponseWriter, request *http.Request) {
	trainers, err := handler.userUseCase.ListTrainers(request.Context())
	if err != nil {
		writeInternalError(writer)
		return
	}

	response := getTrainersResponse{
		Trainers: make([]trainerListItemResponse, 0, len(trainers)),
	}
	for _, trainer := range trainers {
		response.Trainers = append(response.Trainers, handler.newTrainerListItemResponse(trainer))
	}

	writeJSON(writer, http.StatusOK, response)
}

func (handler *Handler) handleGetProfile(writer http.ResponseWriter, request *http.Request) {
	userID, err := strconv.ParseInt(request.PathValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		writeBadRequest(writer)
		return
	}

	user, err := handler.userUseCase.GetByID(request.Context(), userID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			writeNotFound(writer, "Пользователь не найден")
			return
		}

		writeInternalError(writer)
		return
	}

	isMe, err := handler.isCurrentUser(request, userID)
	if err != nil {
		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, handler.newProfileResponse(user, isMe))
}

func (handler *Handler) handlePatchProfileMe(writer http.ResponseWriter, request *http.Request) {
	userID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	command, validationErrors, err := decodeUpdateProfileRequest(request)
	if err != nil {
		writeBadRequest(writer)
		return
	}

	if len(validationErrors) > 0 {
		writeValidationError(writer, validationErrors)
		return
	}

	user, err := handler.userUseCase.UpdateProfile(request.Context(), userID, command)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			writeUnauthorized(writer)
			return
		case errors.Is(err, usecase.ErrUsernameExists):
			writeConflict(writer, "username_exists", "Username уже существует")
			return
		case errors.Is(err, usecase.ErrTrainerProfileForbidden):
			writeForbidden(writer, "Только тренер может обновлять trainer_details")
			return
		case errors.Is(err, usecase.ErrSportTypeNotFound):
			writeValidationError(writer, []validationErrorField{{
				Field:   "trainer_details.sports",
				Message: "Указан несуществующий вид спорта",
			}})
			return
		default:
			writeInternalError(writer)
			return
		}
	}

	writeJSON(writer, http.StatusOK, handler.newProfileResponse(user, true))
}

func (handler *Handler) handlePostProfileAvatar(writer http.ResponseWriter, request *http.Request) {
	userID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maxAvatarSize+1024)
	if err := request.ParseMultipartForm(maxAvatarSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			writeValidationError(writer, []validationErrorField{{
				Field:   "avatar",
				Message: "Размер файла не должен превышать 5 MB",
			}})
			return
		}

		writeBadRequest(writer)
		return
	}
	if request.MultipartForm != nil {
		defer request.MultipartForm.RemoveAll()
	}

	file, fileHeader, err := request.FormFile("avatar")
	if err != nil {
		writeValidationError(writer, []validationErrorField{{
			Field:   "avatar",
			Message: "Нужно приложить файл аватарки",
		}})
		return
	}
	defer file.Close()

	content, contentType, validationErrors, err := decodeAvatarFile(file)
	if err != nil {
		writeInternalError(writer)
		return
	}
	if len(validationErrors) > 0 {
		writeValidationError(writer, validationErrors)
		return
	}

	user, err := handler.userUseCase.UploadAvatar(
		request.Context(),
		userID,
		fileHeader.Filename,
		contentType,
		bytes.NewReader(content),
		int64(len(content)),
	)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			writeUnauthorized(writer)
			return
		default:
			writeInternalError(writer)
			return
		}
	}

	if user.AvatarURL == nil {
		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, avatarUploadResponse{
		AvatarURL: *user.AvatarURL,
	})
}

func (handler *Handler) handleDeleteProfileAvatar(writer http.ResponseWriter, request *http.Request) {
	userID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	err := handler.userUseCase.DeleteAvatar(request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			writeUnauthorized(writer)
			return
		default:
			writeInternalError(writer)
			return
		}
	}

	writeNoContent(writer)
}

func (handler *Handler) newProfileResponse(user domain.User, isMe bool) profileResponse {
	trainerDetails := newTrainerDetailsResponse(user.TrainerDetails)

	return profileResponse{
		UserID:         user.ID,
		IsMe:           isMe,
		IsTrainer:      user.IsTrainer,
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Bio:            user.Bio,
		AvatarURL:      handler.normalizePublicURL(user.AvatarURL),
		TrainerDetails: trainerDetails,
	}
}

func (handler *Handler) newTrainerListItemResponse(trainer domain.TrainerListItem) trainerListItemResponse {
	return trainerListItemResponse{
		UserID:         trainer.ID,
		IsTrainer:      true,
		Username:       trainer.Username,
		FirstName:      trainer.FirstName,
		LastName:       trainer.LastName,
		Bio:            trainer.Bio,
		AvatarURL:      handler.normalizePublicURL(trainer.AvatarURL),
		TrainerDetails: newTrainerDetailsResponse(trainer.TrainerDetails),
	}
}

func newTrainerDetailsResponse(details *domain.TrainerDetails) *trainerDetailsResponse {
	if details == nil {
		return nil
	}

	sports := make([]trainerSportResponse, 0, len(details.Sports))
	for _, sport := range details.Sports {
		sports = append(sports, trainerSportResponse{
			SportTypeID:     sport.SportTypeID,
			ExperienceYears: sport.ExperienceYears,
			SportsRank:      sport.SportsRank,
		})
	}

	return &trainerDetailsResponse{
		EducationDegree: details.EducationDegree,
		CareerSinceDate: details.CareerSinceDate.Format("2006-01-02"),
		Sports:          sports,
	}
}

func decodeUpdateProfileRequest(request *http.Request) (usecase.UpdateProfileCommand, []validationErrorField, error) {
	var raw map[string]json.RawMessage

	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&raw); err != nil {
		return usecase.UpdateProfileCommand{}, nil, err
	}

	command := usecase.UpdateProfileCommand{}
	validationErrors := make([]validationErrorField, 0)

	if len(raw) == 0 {
		return command, []validationErrorField{{
			Field:   "body",
			Message: "Нужно указать хотя бы одно поле для обновления",
		}}, nil
	}

	for field := range raw {
		switch field {
		case "username", "first_name", "last_name", "bio", "trainer_details":
		default:
			return usecase.UpdateProfileCommand{}, nil, errors.New("unknown field")
		}
	}

	if rawValue, ok := raw["username"]; ok {
		var username string
		if err := json.Unmarshal(rawValue, &username); err != nil {
			return usecase.UpdateProfileCommand{}, nil, err
		}

		command.HasUsername = true
		command.Username = username
		if !usernamePattern.MatchString(username) {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   "username",
				Message: "Username должен содержать от 3 до 30 символов и только буквы, цифры или _",
			})
		}
	}

	if rawValue, ok := raw["first_name"]; ok {
		var firstName string
		if err := json.Unmarshal(rawValue, &firstName); err != nil {
			return usecase.UpdateProfileCommand{}, nil, err
		}

		command.HasFirstName = true
		command.FirstName = firstName
		if len(firstName) < 1 || len(firstName) > 100 {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   "first_name",
				Message: "First name должен содержать от 1 до 100 символов",
			})
		}
	}

	if rawValue, ok := raw["last_name"]; ok {
		var lastName string
		if err := json.Unmarshal(rawValue, &lastName); err != nil {
			return usecase.UpdateProfileCommand{}, nil, err
		}

		command.HasLastName = true
		command.LastName = lastName
		if len(lastName) < 1 || len(lastName) > 100 {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   "last_name",
				Message: "Last name должен содержать от 1 до 100 символов",
			})
		}
	}

	if rawValue, ok := raw["bio"]; ok {
		command.HasBio = true

		if !bytes.Equal(bytes.TrimSpace(rawValue), []byte("null")) {
			var bio string
			if err := json.Unmarshal(rawValue, &bio); err != nil {
				return usecase.UpdateProfileCommand{}, nil, err
			}

			command.Bio = &bio
			if len(bio) > 1000 {
				validationErrors = append(validationErrors, validationErrorField{
					Field:   "bio",
					Message: "Bio должен содержать не более 1000 символов",
				})
			}
		}
	}

	if rawValue, ok := raw["trainer_details"]; ok {
		var trainerDetailsRaw map[string]json.RawMessage
		if err := json.Unmarshal(rawValue, &trainerDetailsRaw); err != nil {
			return usecase.UpdateProfileCommand{}, nil, err
		}

		for field := range trainerDetailsRaw {
			switch field {
			case "education_degree", "career_since_date", "sports":
			default:
				return usecase.UpdateProfileCommand{}, nil, errors.New("unknown trainer_details field")
			}
		}

		if rawEducationDegree, ok := trainerDetailsRaw["education_degree"]; ok {
			command.HasEducationDegree = true
			if !bytes.Equal(bytes.TrimSpace(rawEducationDegree), []byte("null")) {
				var educationDegree string
				if err := json.Unmarshal(rawEducationDegree, &educationDegree); err != nil {
					return usecase.UpdateProfileCommand{}, nil, err
				}

				command.EducationDegree = &educationDegree
				if len(educationDegree) > 255 {
					validationErrors = append(validationErrors, validationErrorField{
						Field:   "trainer_details.education_degree",
						Message: "Образование должно содержать не более 255 символов",
					})
				}
			}
		}

		if rawCareerSinceDate, ok := trainerDetailsRaw["career_since_date"]; ok {
			command.HasCareerSinceDate = true

			var careerSinceDateRaw string
			if err := json.Unmarshal(rawCareerSinceDate, &careerSinceDateRaw); err != nil {
				return usecase.UpdateProfileCommand{}, nil, err
			}

			careerSinceDate, err := time.Parse("2006-01-02", careerSinceDateRaw)
			if err != nil {
				validationErrors = append(validationErrors, validationErrorField{
					Field:   "trainer_details.career_since_date",
					Message: "Неверный формат даты",
				})
			} else if careerSinceDate.After(time.Now()) {
				validationErrors = append(validationErrors, validationErrorField{
					Field:   "trainer_details.career_since_date",
					Message: "Дата начала карьеры не может быть в будущем",
				})
			} else {
				command.CareerSinceDate = careerSinceDate
			}
		}

		if rawSports, ok := trainerDetailsRaw["sports"]; ok {
			var sports []trainerRegisterSportRequest
			if err := json.Unmarshal(rawSports, &sports); err != nil {
				return usecase.UpdateProfileCommand{}, nil, err
			}

			command.HasSports = true
			command.Sports = make([]usecase.RegisterTrainerSportCommand, 0, len(sports))

			if len(sports) == 0 {
				validationErrors = append(validationErrors, validationErrorField{
					Field:   "trainer_details.sports",
					Message: "Нужно указать хотя бы один вид спорта",
				})
			}

			seenSportTypeIDs := make(map[int64]struct{}, len(sports))
			for _, sport := range sports {
				command.Sports = append(command.Sports, usecase.RegisterTrainerSportCommand{
					SportTypeID:     sport.SportTypeID,
					ExperienceYears: sport.ExperienceYears,
					SportsRank:      sport.SportsRank,
				})

				if sport.SportTypeID <= 0 {
					validationErrors = append(validationErrors, validationErrorField{
						Field:   "trainer_details.sports",
						Message: "sport_type_id должен быть положительным числом",
					})
				}

				if sport.ExperienceYears < 0 {
					validationErrors = append(validationErrors, validationErrorField{
						Field:   "trainer_details.sports",
						Message: "experience_years не может быть отрицательным",
					})
				}

				if sport.SportsRank != nil && len(*sport.SportsRank) > 100 {
					validationErrors = append(validationErrors, validationErrorField{
						Field:   "trainer_details.sports",
						Message: "sports_rank должен содержать не более 100 символов",
					})
				}

				if _, ok := seenSportTypeIDs[sport.SportTypeID]; ok {
					validationErrors = append(validationErrors, validationErrorField{
						Field:   "trainer_details.sports",
						Message: "sport_type_id не должен повторяться",
					})
				}
				seenSportTypeIDs[sport.SportTypeID] = struct{}{}
			}
		}
	}

	return command, validationErrors, nil
}

func decodeAvatarFile(file io.Reader) ([]byte, string, []validationErrorField, error) {
	content, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1))
	if err != nil {
		return nil, "", nil, err
	}

	if len(content) == 0 {
		return nil, "", []validationErrorField{{
			Field:   "avatar",
			Message: "Файл аватарки не должен быть пустым",
		}}, nil
	}

	if len(content) > maxAvatarSize {
		return nil, "", []validationErrorField{{
			Field:   "avatar",
			Message: "Размер файла не должен превышать 5 MB",
		}}, nil
	}

	contentType := http.DetectContentType(content)
	if _, ok := allowedAvatarContentTypes[contentType]; !ok {
		return nil, "", []validationErrorField{{
			Field:   "avatar",
			Message: "Поддерживаются только JPEG, PNG и WEBP",
		}}, nil
	}

	return content, contentType, nil, nil
}

func (handler *Handler) isCurrentUser(request *http.Request, userID int64) (bool, error) {
	currentUserID, err := handler.currentUserID(request)
	if err != nil {
		return false, err
	}

	return currentUserID == userID, nil
}

func (handler *Handler) currentUserID(request *http.Request) (int64, error) {
	cookie, err := request.Cookie(handler.authCookieName)
	if err != nil {
		return 0, nil
	}

	currentUserID, err := handler.sessionUseCase.GetUserIDBySessionID(request.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, usecase.ErrSessionNotFound) {
			return 0, nil
		}

		return 0, err
	}

	return currentUserID, nil
}
