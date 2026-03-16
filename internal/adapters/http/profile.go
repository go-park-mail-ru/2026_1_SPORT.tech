package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

type profileResponse struct {
	UserID    int64   `json:"user_id"`
	IsMe      bool    `json:"is_me"`
	IsTrainer bool    `json:"is_trainer"`
	Username  string  `json:"username"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Bio       *string `json:"bio"`
	AvatarURL *string `json:"avatar_url"`
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

	writeJSON(writer, http.StatusOK, profileResponse{
		UserID:    user.ID,
		IsMe:      isMe,
		IsTrainer: user.IsTrainer,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Bio:       user.Bio,
		AvatarURL: user.AvatarURL,
	})
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
