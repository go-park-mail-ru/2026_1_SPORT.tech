package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

type profilePostsResponse struct {
	UserID int64             `json:"user_id"`
	Posts  []postListItemDTO `json:"posts"`
}

type postListItemDTO struct {
	PostID     int64     `json:"post_id"`
	TrainerID  int64     `json:"trainer_id"`
	MinTierID  *int64    `json:"min_tier_id"`
	Title      string    `json:"title"`
	CreatedAt  time.Time `json:"created_at"`
	CanView    bool      `json:"can_view"`
	LikesCount int64     `json:"likes_count"`
	IsLiked    bool      `json:"is_liked"`
}

func (handler *Handler) handleGetProfilePosts(writer http.ResponseWriter, request *http.Request) {
	userID, err := strconv.ParseInt(request.PathValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		writeBadRequest(writer)
		return
	}

	_, err = handler.userUseCase.GetByID(request.Context(), userID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			writeNotFound(writer, "Пользователь не найден")
			return
		}

		writeInternalError(writer)
		return
	}

	currentUserID, err := handler.currentUserID(request)
	if err != nil {
		writeInternalError(writer)
		return
	}

	posts, err := handler.postUseCase.ListProfilePosts(request.Context(), userID, currentUserID)
	if err != nil {
		writeInternalError(writer)
		return
	}

	response := profilePostsResponse{
		UserID: userID,
		Posts:  make([]postListItemDTO, 0, len(posts)),
	}

	for _, post := range posts {
		response.Posts = append(response.Posts, postListItemDTO{
			PostID:     post.PostID,
			TrainerID:  post.TrainerID,
			MinTierID:  post.MinTierID,
			Title:      post.Title,
			CreatedAt:  post.CreatedAt,
			CanView:    post.CanView,
			LikesCount: post.LikesCount,
			IsLiked:    post.IsLiked,
		})
	}

	writeJSON(writer, http.StatusOK, response)
}
