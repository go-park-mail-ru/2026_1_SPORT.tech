package handler

import (
	"errors"
	nethttp "net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
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

func (handler *Handler) handleGetPost(writer nethttp.ResponseWriter, request *nethttp.Request) {
	postID, err := strconv.ParseInt(request.PathValue("post_id"), 10, 64)
	if err != nil || postID <= 0 {
		writeBadRequest(writer)
		return
	}

	currentUserID, err := handler.currentUserID(request)
	if err != nil {
		writeInternalError(writer)
		return
	}

	post, err := handler.postService.GetByID(request.Context(), postID, currentUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrPostNotFound):
			writeNotFound(writer, "Пост не найден")
			return
		case errors.Is(err, service.ErrPostForbidden):
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

	writeJSON(writer, nethttp.StatusOK, postResponse{
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
