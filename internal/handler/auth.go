package handler

import (
	"context"
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

	writeJSON(writer, nethttp.StatusOK, authResponse{
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
	})
}
