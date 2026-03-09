package handler

import (
	"errors"
	nethttp "net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
)

type profileResponse struct {
	UserID    int64               `json:"user_id"`
	IsMe      bool                `json:"is_me"`
	IsTrainer bool                `json:"is_trainer"`
	Profile   userProfileResponse `json:"profile"`
}

func (handler *Handler) handleGetProfile(writer nethttp.ResponseWriter, request *nethttp.Request) {
	userID, err := strconv.ParseInt(request.PathValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		writeBadRequest(writer)
		return
	}

	user, err := handler.userService.GetByID(request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
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

	writeJSON(writer, nethttp.StatusOK, profileResponse{
		UserID:    user.ID,
		IsMe:      isMe,
		IsTrainer: user.IsTrainer,
		Profile: userProfileResponse{
			Username:  user.Profile.Username,
			FirstName: user.Profile.FirstName,
			LastName:  user.Profile.LastName,
			Bio:       user.Profile.Bio,
			AvatarURL: user.Profile.AvatarURL,
		},
	})
}

func (handler *Handler) isCurrentUser(request *nethttp.Request, userID int64) (bool, error) {
	cookie, err := request.Cookie(handler.authCookieName)
	if err != nil {
		return false, nil
	}

	currentUserID, err := handler.sessionService.GetUserIDBySessionID(request.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, service.ErrSessionNotFound) {
			return false, nil
		}

		return false, err
	}

	return currentUserID == userID, nil
}
