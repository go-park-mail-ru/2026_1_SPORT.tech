package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

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
	var trainerDetails *trainerDetailsResponse
	if user.TrainerDetails != nil {
		sports := make([]trainerSportResponse, 0, len(user.TrainerDetails.Sports))
		for _, sport := range user.TrainerDetails.Sports {
			sports = append(sports, trainerSportResponse{
				SportTypeID:     sport.SportTypeID,
				ExperienceYears: sport.ExperienceYears,
				SportsRank:      sport.SportsRank,
			})
		}

		trainerDetails = &trainerDetailsResponse{
			EducationDegree: user.TrainerDetails.EducationDegree,
			CareerSinceDate: user.TrainerDetails.CareerSinceDate.Format("2006-01-02"),
			Sports:          sports,
		}
	}

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
		case "username", "first_name", "last_name", "bio":
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
