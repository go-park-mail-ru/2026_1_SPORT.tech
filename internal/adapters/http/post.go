package handler

import (
	"encoding/json"
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

type createPostAttachmentRequest struct {
	Kind    string `json:"kind"`
	FileURL string `json:"file_url"`
}

type createPostRequest struct {
	MinTierID   *int64                       `json:"min_tier_id"`
	Title       string                       `json:"title"`
	TextContent string                       `json:"text_content"`
	Attachments []createPostAttachmentRequest `json:"attachments"`
}

func (handler *Handler) handleGetPost(writer http.ResponseWriter, request *http.Request) {
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

	writeJSON(writer, http.StatusOK, newPostResponse(post))
}

func (handler *Handler) handlePostCreate(writer http.ResponseWriter, request *http.Request) {
	currentUserID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	user, err := handler.userUseCase.GetByID(request.Context(), currentUserID)
	if err != nil {
		if errors.Is(err, usecase.ErrUserNotFound) {
			writeUnauthorized(writer)
			return
		}

		writeInternalError(writer)
		return
	}

	if !user.IsTrainer {
		writeForbidden(writer, "Только тренер может создавать посты")
		return
	}

	var createRequest createPostRequest

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&createRequest); err != nil {
		writeBadRequest(writer)
		return
	}

	validationErrors := validateCreatePostRequest(createRequest)
	if len(validationErrors) > 0 {
		writeValidationError(writer, validationErrors)
		return
	}

	attachments := make([]usecase.CreatePostAttachmentCommand, 0, len(createRequest.Attachments))
	for _, attachment := range createRequest.Attachments {
		attachments = append(attachments, usecase.CreatePostAttachmentCommand{
			Kind:    attachment.Kind,
			FileURL: attachment.FileURL,
		})
	}

	post, err := handler.postUseCase.Create(request.Context(), currentUserID, usecase.CreatePostCommand{
		MinTierID:   createRequest.MinTierID,
		Title:       createRequest.Title,
		TextContent: createRequest.TextContent,
		Attachments: attachments,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrPostTierNotFound) {
			writeValidationError(writer, []validationErrorField{{
				Field:   "min_tier_id",
				Message: "Указан несуществующий tier или tier не принадлежит тренеру",
			}})
			return
		}

		writeInternalError(writer)
		return
	}

	writeJSON(writer, http.StatusCreated, newPostResponse(post))
}

func newPostResponse(post domain.Post) postResponse {
	attachments := make([]postAttachmentResponse, 0, len(post.Attachments))
	for _, attachment := range post.Attachments {
		attachments = append(attachments, postAttachmentResponse{
			PostAttachmentID: attachment.PostAttachmentID,
			Kind:             attachment.Kind,
			FileURL:          attachment.FileURL,
		})
	}

	return postResponse{
		PostID:      post.PostID,
		TrainerID:   post.TrainerID,
		MinTierID:   post.MinTierID,
		Title:       post.Title,
		TextContent: post.TextContent,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Attachments: attachments,
	}
}

func validateCreatePostRequest(request createPostRequest) []validationErrorField {
	validationErrors := make([]validationErrorField, 0)

	if request.MinTierID != nil && *request.MinTierID <= 0 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "min_tier_id",
			Message: "min_tier_id должен быть положительным числом",
		})
	}

	if len(request.Title) < 1 || len(request.Title) > 200 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "title",
			Message: "Title должен содержать от 1 до 200 символов",
		})
	}

	if len(request.TextContent) < 1 || len(request.TextContent) > 10000 {
		validationErrors = append(validationErrors, validationErrorField{
			Field:   "text_content",
			Message: "Text content должен содержать от 1 до 10000 символов",
		})
	}

	for index, attachment := range request.Attachments {
		fieldPrefix := "attachments[" + strconv.Itoa(index) + "]"

		switch attachment.Kind {
		case "image", "video", "document":
		default:
			validationErrors = append(validationErrors, validationErrorField{
				Field:   fieldPrefix + ".kind",
				Message: "kind должен быть одним из: image, video, document",
			})
		}

		if attachment.FileURL == "" {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   fieldPrefix + ".file_url",
				Message: "file_url обязателен",
			})
		}
	}

	return validationErrors
}
