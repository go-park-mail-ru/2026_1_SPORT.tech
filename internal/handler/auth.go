package handler

import (
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"regexp"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,30}$`)
var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type sessionService interface {
	CreateSession(ctx context.Context, userID int64) (string, error)
	GetUserIDBySessionID(ctx context.Context, sessionID string) (int64, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type userService interface {
	GetByID(ctx context.Context, userID int64) (service.User, error)
	RegisterClient(ctx context.Context, params service.RegisterClientParams) (service.User, error)
	RegisterTrainer(ctx context.Context, params service.RegisterTrainerParams) (service.User, error)
	Authenticate(ctx context.Context, email string, password string) (service.User, error)
}

type userProfileResponse struct {
	Username  string  `json:"username"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
}

type userResponse struct {
	UserID    int64               `json:"user_id"`
	Username  string              `json:"username"`
	Email     string              `json:"email"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	IsTrainer bool                `json:"is_trainer"`
	IsAdmin   bool                `json:"is_admin"`
	Profile   userProfileResponse `json:"profile"`
}

type authResponse struct {
	User userResponse `json:"user"`
}

type clientRegisterRequest struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
}

type trainerRegisterSportRequest struct {
	SportTypeID     int64   `json:"sport_type_id"`
	ExperienceYears int     `json:"experience_years"`
	SportsRank      *string `json:"sports_rank"`
}

type trainerRegisterDetailsRequest struct {
	EducationDegree *string                       `json:"education_degree"`
	CareerSinceDate string                        `json:"career_since_date"`
	Sports          []trainerRegisterSportRequest `json:"sports"`
}

type trainerRegisterRequest struct {
	Username       string                        `json:"username"`
	Email          string                        `json:"email"`
	Password       string                        `json:"password"`
	PasswordRepeat string                        `json:"password_repeat"`
	FirstName      string                        `json:"first_name"`
	LastName       string                        `json:"last_name"`
	TrainerDetails trainerRegisterDetailsRequest `json:"trainer_details"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (handler *Handler) AuthMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(writer nethttp.ResponseWriter, request *nethttp.Request) {
		cookie, err := request.Cookie(handler.authCookieName)
		if err != nil {
			writeUnauthorized(writer)
			return
		}

		userID, err := handler.sessionService.GetUserIDBySessionID(request.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, service.ErrSessionNotFound) {
				writeUnauthorized(writer)
				return
			}

			writeInternalError(writer)
			return
		}

		ctx := context.WithValue(request.Context(), userIDContextKey, userID)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func (handler *Handler) setSessionCookie(writer nethttp.ResponseWriter, sessionID string) {
	nethttp.SetCookie(writer, &nethttp.Cookie{
		Name:     handler.authCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: nethttp.SameSiteLaxMode,
	})
}

func (handler *Handler) clearSessionCookie(writer nethttp.ResponseWriter) {
	nethttp.SetCookie(writer, &nethttp.Cookie{
		Name:     handler.authCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: nethttp.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func userIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDContextKey).(int64)
	return userID, ok
}

func (handler *Handler) handlePostAuthRegisterClient(writer nethttp.ResponseWriter, request *nethttp.Request) {
	var registerRequest clientRegisterRequest

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registerRequest); err != nil {
		writeBadRequest(writer)
		return
	}

	validationErrors := validateClientRegisterRequest(registerRequest)
	if len(validationErrors) > 0 {
		writeValidationError(writer, validationErrors)
		return
	}

	user, err := handler.userService.RegisterClient(request.Context(), service.RegisterClientParams{
		Username:  registerRequest.Username,
		Email:     registerRequest.Email,
		Password:  registerRequest.Password,
		FirstName: registerRequest.FirstName,
		LastName:  registerRequest.LastName,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			writeConflict(writer, "email_exists", "Email уже существует")
			return
		case errors.Is(err, service.ErrUsernameExists):
			writeConflict(writer, "username_exists", "Username уже существует")
			return
		default:
			writeInternalError(writer)
			return
		}
	}

	sessionID, err := handler.sessionService.CreateSession(request.Context(), user.ID)
	if err != nil {
		writeInternalError(writer)
		return
	}

	handler.setSessionCookie(writer, sessionID)
	writeJSON(writer, nethttp.StatusCreated, newAuthResponse(user))
}

func (handler *Handler) handlePostAuthRegisterTrainer(writer nethttp.ResponseWriter, request *nethttp.Request) {
	var registerRequest trainerRegisterRequest

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registerRequest); err != nil {
		writeBadRequest(writer)
		return
	}

	careerSinceDate, validationErrors := validateTrainerRegisterRequest(registerRequest)
	if len(validationErrors) > 0 {
		writeValidationError(writer, validationErrors)
		return
	}

	sports := make([]service.RegisterTrainerSportParams, 0, len(registerRequest.TrainerDetails.Sports))
	for _, sport := range registerRequest.TrainerDetails.Sports {
		sports = append(sports, service.RegisterTrainerSportParams{
			SportTypeID:     sport.SportTypeID,
			ExperienceYears: sport.ExperienceYears,
			SportsRank:      sport.SportsRank,
		})
	}

	user, err := handler.userService.RegisterTrainer(request.Context(), service.RegisterTrainerParams{
		Username:        registerRequest.Username,
		Email:           registerRequest.Email,
		Password:        registerRequest.Password,
		FirstName:       registerRequest.FirstName,
		LastName:        registerRequest.LastName,
		EducationDegree: registerRequest.TrainerDetails.EducationDegree,
		CareerSinceDate: careerSinceDate,
		Sports:          sports,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailExists):
			writeConflict(writer, "email_exists", "Email уже существует")
			return
		case errors.Is(err, service.ErrUsernameExists):
			writeConflict(writer, "username_exists", "Username уже существует")
			return
		case errors.Is(err, service.ErrSportTypeNotFound):
			writeValidationError(writer, []validationErrorField{{
				Field:   "trainer_details.sports",
				Message: "Указан несуществующий sport_type_id",
			}})
			return
		default:
			writeInternalError(writer)
			return
		}
	}

	sessionID, err := handler.sessionService.CreateSession(request.Context(), user.ID)
	if err != nil {
		writeInternalError(writer)
		return
	}

	handler.setSessionCookie(writer, sessionID)
	writeJSON(writer, nethttp.StatusCreated, newAuthResponse(user))
}

func (handler *Handler) handlePostAuthLogin(writer nethttp.ResponseWriter, request *nethttp.Request) {
	var loginRequest loginRequest

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&loginRequest); err != nil {
		writeBadRequest(writer)
		return
	}

	if loginRequest.Email == "" || loginRequest.Password == "" {
		writeBadRequest(writer)
		return
	}

	user, err := handler.userService.Authenticate(request.Context(), loginRequest.Email, loginRequest.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeInvalidCredentials(writer)
			return
		}

		writeInternalError(writer)
		return
	}

	sessionID, err := handler.sessionService.CreateSession(request.Context(), user.ID)
	if err != nil {
		writeInternalError(writer)
		return
	}

	handler.setSessionCookie(writer, sessionID)
	writeJSON(writer, nethttp.StatusOK, newAuthResponse(user))
}

func (handler *Handler) handleGetAuthMe(writer nethttp.ResponseWriter, request *nethttp.Request) {
	userID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	user, err := handler.userService.GetByID(request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeUnauthorized(writer)
			return
		}

		writeInternalError(writer)
		return
	}

	writeJSON(writer, nethttp.StatusOK, newAuthResponse(user))
}

func (handler *Handler) handlePostAuthLogout(writer nethttp.ResponseWriter, request *nethttp.Request) {
	cookie, err := request.Cookie(handler.authCookieName)
	if err != nil {
		writeUnauthorized(writer)
		return
	}

	if err := handler.sessionService.RevokeSession(request.Context(), cookie.Value); err != nil {
		writeInternalError(writer)
		return
	}

	handler.clearSessionCookie(writer)
	writeNoContent(writer)
}

func newAuthResponse(user service.User) authResponse {
	return authResponse{
		User: userResponse{
			UserID:    user.ID,
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			IsTrainer: user.IsTrainer,
			IsAdmin:   user.IsAdmin,
			Profile: userProfileResponse{
				Username:  user.Profile.Username,
				FirstName: user.Profile.FirstName,
				LastName:  user.Profile.LastName,
				Bio:       user.Profile.Bio,
				AvatarURL: user.Profile.AvatarURL,
			},
		},
	}
}

func validateClientRegisterRequest(request clientRegisterRequest) []validationErrorField {
	validationErrors := make([]validationErrorField, 0)

	if !usernamePattern.MatchString(request.Username) {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "username",
			Message: "Username должен содержать от 3 до 30 символов и только буквы, цифры или _",
		})
	}

	if !emailPattern.MatchString(request.Email) || len(request.Email) > 254 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "email",
			Message: "Неверный формат email",
		})
	}

	if len(request.Password) < 8 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "password",
			Message: "Пароль должен содержать минимум 8 символов",
		})
	}

	if request.Password != request.PasswordRepeat {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "password_repeat",
			Message: "Пароли не совпадают",
		})
	}

	if len(request.FirstName) < 1 || len(request.FirstName) > 100 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "first_name",
			Message: "Имя должно содержать от 1 до 100 символов",
		})
	}

	if len(request.LastName) < 1 || len(request.LastName) > 100 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "last_name",
			Message: "Фамилия должна содержать от 1 до 100 символов",
		})
	}

	return validationErrors
}

func validateTrainerRegisterRequest(request trainerRegisterRequest) (time.Time, []validationErrorField) {
	validationErrors := validateClientRegisterRequest(clientRegisterRequest{
		Username:       request.Username,
		Email:          request.Email,
		Password:       request.Password,
		PasswordRepeat: request.PasswordRepeat,
		FirstName:      request.FirstName,
		LastName:       request.LastName,
	})

	if request.TrainerDetails.EducationDegree != nil && len(*request.TrainerDetails.EducationDegree) > 255 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "trainer_details.education_degree",
			Message: "Образование должно содержать не более 255 символов",
		})
	}

	careerSinceDate, err := time.Parse("2006-01-02", request.TrainerDetails.CareerSinceDate)
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
	}

	if len(request.TrainerDetails.Sports) == 0 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "trainer_details.sports",
			Message: "Нужно указать хотя бы один вид спорта",
		})
	}

	seenSportTypeIDs := make(map[int64]struct{}, len(request.TrainerDetails.Sports))
	for index, sport := range request.TrainerDetails.Sports {
		fieldPrefix := "trainer_details.sports"

		if sport.SportTypeID <= 0 {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   fieldPrefix,
				Message: "sport_type_id должен быть положительным числом",
			})
		}

		if sport.ExperienceYears < 0 {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   fieldPrefix,
				Message: "experience_years не может быть отрицательным",
			})
		}

		if sport.SportsRank != nil && len(*sport.SportsRank) > 100 {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   fieldPrefix,
				Message: "sports_rank должен содержать не более 100 символов",
			})
		}

		if _, ok := seenSportTypeIDs[sport.SportTypeID]; ok {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   fieldPrefix,
				Message: "sport_type_id не должен повторяться",
			})
		}
		seenSportTypeIDs[sport.SportTypeID] = struct{}{}

		_ = index
	}

	return careerSinceDate, validationErrors
}
