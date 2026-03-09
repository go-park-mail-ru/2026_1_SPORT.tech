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
