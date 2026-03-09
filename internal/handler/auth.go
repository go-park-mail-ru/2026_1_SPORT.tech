package handler

import (
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

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
