package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
)

type postAttachmentResponse struct {
	PostAttachmentID int64  `json:"post_attachment_id"`
	Kind             string `json:"kind"`
	FileURL          string `json:"file_url"`
}

type postResponse struct {
	PostID      int64                    `json:"post_id"`
	TrainerID   int64                    `json:"trainer_id"`
	MinTierID   *int64                   `json:"min_tier_id"`
	Title       string                   `json:"title"`
	TextContent string                   `json:"text_content"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
	Attachments []postAttachmentResponse `json:"attachments"`
}

type postLikeResponse struct {
	PostID     int64 `json:"post_id"`
	LikesCount int64 `json:"likes_count"`
	IsLiked    bool  `json:"is_liked"`
}

func (handler *Handler) handleGetPost(writer http.ResponseWriter, request *http.Request) {
	postID, ok := parsePostID(request)
	if !ok {
		writeBadRequest(writer)
		return
	}

	currentUserID, err := handler.currentUserID(request)
	if err != nil {
		writeInternalError(writer)
		return
	}

	post, err := handler.postUseCase.GetByID(request.Context(), postID, currentUserID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrPostNotFound):
			writeNotFound(writer, "Пост не найден")
			return
		case errors.Is(err, usecase.ErrPostForbidden):
			writeForbidden(writer, "Нет доступа к этому посту")
			return
		default:
			writeInternalError(writer)
			return
		}
	}

	attachments := make([]postAttachmentResponse, 0, len(post.Attachments))
	for _, attachment := range post.Attachments {
		attachments = append(attachments, postAttachmentResponse{
			PostAttachmentID: attachment.PostAttachmentID,
			Kind:             attachment.Kind,
			FileURL:          attachment.FileURL,
		})
	}

	writeJSON(writer, http.StatusOK, postResponse{
		PostID:      post.PostID,
		TrainerID:   post.TrainerID,
		MinTierID:   post.MinTierID,
		Title:       post.Title,
		TextContent: post.TextContent,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Attachments: attachments,
	})
}

func (handler *Handler) handlePostPostLike(writer http.ResponseWriter, request *http.Request) {
	userID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	postID, ok := parsePostID(request)
	if !ok {
		writeBadRequest(writer)
		return
	}

	postLikeStatus, err := handler.postUseCase.SetLike(request.Context(), postID, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrPostNotFound) {
			writeNotFound(writer, "Пост не найден")
			return
		}

		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, newPostLikeResponse(postLikeStatus))
}

func (handler *Handler) handleDeletePostLike(writer http.ResponseWriter, request *http.Request) {
	userID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	postID, ok := parsePostID(request)
	if !ok {
		writeBadRequest(writer)
		return
	}

	postLikeStatus, err := handler.postUseCase.DeleteLike(request.Context(), postID, userID)
	if err != nil {
		if errors.Is(err, usecase.ErrPostNotFound) {
			writeNotFound(writer, "Пост не найден")
			return
		}

		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusOK, newPostLikeResponse(postLikeStatus))
}

func parsePostID(request *http.Request) (int64, bool) {
	postID, err := strconv.ParseInt(request.PathValue("post_id"), 10, 64)
	if err != nil || postID <= 0 {
		return 0, false
	}

	return postID, true
}

func newPostLikeResponse(postLikeStatus domain.PostLikeStatus) postLikeResponse {
	return postLikeResponse{
		PostID:     postLikeStatus.PostID,
		LikesCount: postLikeStatus.LikesCount,
		IsLiked:    postLikeStatus.IsLiked,
	}
}
