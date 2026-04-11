package handler

import (
	"bytes"
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

type postLikeResponse struct {
	PostID     int64 `json:"post_id"`
	LikesCount int64 `json:"likes_count"`
	IsLiked    bool  `json:"is_liked"`
}

type createPostAttachmentRequest struct {
	Kind    string `json:"kind"`
	FileURL string `json:"file_url"`
}

type createPostRequest struct {
	MinTierID   *int64                        `json:"min_tier_id"`
	Title       string                        `json:"title"`
	TextContent string                        `json:"text_content"`
	Attachments []createPostAttachmentRequest `json:"attachments"`
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

	post, err := handler.postUseCase.Create(request.Context(), currentUserID, usecase.CreatePostCommand{
		MinTierID:   createRequest.MinTierID,
		Title:       createRequest.Title,
		TextContent: createRequest.TextContent,
		Attachments: newCreatePostAttachments(createRequest.Attachments),
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

func (handler *Handler) handlePatchPost(writer http.ResponseWriter, request *http.Request) {
	currentUserID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	postID, err := strconv.ParseInt(request.PathValue("post_id"), 10, 64)
	if err != nil || postID <= 0 {
		writeBadRequest(writer)
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
		writeForbidden(writer, "Только тренер может изменять посты")
		return
	}

	command, validationErrors, err := decodeUpdatePostRequest(request)
	if err != nil {
		writeBadRequest(writer)
		return
	}

	if len(validationErrors) > 0 {
		writeValidationError(writer, validationErrors)
		return
	}

	post, err := handler.postUseCase.Update(request.Context(), currentUserID, postID, command)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrPostNotFound):
			writeNotFound(writer, "Пост не найден")
			return
		case errors.Is(err, usecase.ErrPostForbidden):
			writeForbidden(writer, "Нельзя изменять чужой пост")
			return
		case errors.Is(err, usecase.ErrPostTierNotFound):
			writeValidationError(writer, []validationErrorField{{
				Field:   "min_tier_id",
				Message: "Указан несуществующий tier или tier не принадлежит тренеру",
			}})
			return
		default:
			writeInternalError(writer)
			return
		}
	}

	writeJSON(writer, http.StatusOK, newPostResponse(post))
}

func (handler *Handler) handleDeletePost(writer http.ResponseWriter, request *http.Request) {
	currentUserID, ok := userIDFromContext(request.Context())
	if !ok {
		writeInternalError(writer)
		return
	}

	postID, err := strconv.ParseInt(request.PathValue("post_id"), 10, 64)
	if err != nil || postID <= 0 {
		writeBadRequest(writer)
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
		writeForbidden(writer, "Только тренер может удалять посты")
		return
	}

	err = handler.postUseCase.Delete(request.Context(), currentUserID, postID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrPostNotFound):
			writeNotFound(writer, "Пост не найден")
			return
		case errors.Is(err, usecase.ErrPostForbidden):
			writeForbidden(writer, "Нельзя удалять чужой пост")
			return
		default:
			writeInternalError(writer)
			return
		}
	}

	writeNoContent(writer)
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

func decodeUpdatePostRequest(request *http.Request) (usecase.UpdatePostCommand, []validationErrorField, error) {
	var raw map[string]json.RawMessage

	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&raw); err != nil {
		return usecase.UpdatePostCommand{}, nil, err
	}

	command := usecase.UpdatePostCommand{}
	validationErrors := make([]validationErrorField, 0)

	if len(raw) == 0 {
		return command, []validationErrorField{{
			Field:   "body",
			Message: "Нужно указать хотя бы одно поле для обновления",
		}}, nil
	}

	for field := range raw {
		switch field {
		case "min_tier_id", "title", "text_content", "attachments":
		default:
			return usecase.UpdatePostCommand{}, nil, errors.New("unknown field")
		}
	}

	if rawValue, ok := raw["min_tier_id"]; ok {
		command.HasMinTierID = true

		if !bytes.Equal(bytes.TrimSpace(rawValue), []byte("null")) {
			var minTierID int64
			if err := json.Unmarshal(rawValue, &minTierID); err != nil {
				return usecase.UpdatePostCommand{}, nil, err
			}

			command.MinTierID = &minTierID
			if minTierID <= 0 {
				validationErrors = append(validationErrors, validationErrorField{
					Field:   "min_tier_id",
					Message: "min_tier_id должен быть положительным числом",
				})
			}
		}
	}

	if rawValue, ok := raw["title"]; ok {
		var title string
		if err := json.Unmarshal(rawValue, &title); err != nil {
			return usecase.UpdatePostCommand{}, nil, err
		}

		command.Title = &title
		if len(title) < 1 || len(title) > 200 {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   "title",
				Message: "Title должен содержать от 1 до 200 символов",
			})
		}
	}

	if rawValue, ok := raw["text_content"]; ok {
		var textContent string
		if err := json.Unmarshal(rawValue, &textContent); err != nil {
			return usecase.UpdatePostCommand{}, nil, err
		}

		command.TextContent = &textContent
		if len(textContent) < 1 || len(textContent) > 10000 {
			validationErrors = append(validationErrors, validationErrorField{
				Field:   "text_content",
				Message: "Text content должен содержать от 1 до 10000 символов",
			})
		}
	}

	if rawValue, ok := raw["attachments"]; ok {
		if bytes.Equal(bytes.TrimSpace(rawValue), []byte("null")) {
			return usecase.UpdatePostCommand{}, nil, errors.New("attachments must be array")
		}

		var attachments []createPostAttachmentRequest
		if err := json.Unmarshal(rawValue, &attachments); err != nil {
			return usecase.UpdatePostCommand{}, nil, err
		}

		command.HasAttachments = true
		command.Attachments = newCreatePostAttachments(attachments)
		validationErrors = append(validationErrors, validatePostAttachments(attachments)...)
	}

	return command, validationErrors, nil
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

	return append(validationErrors, validatePostAttachments(request.Attachments)...)
}

func validatePostAttachments(attachments []createPostAttachmentRequest) []validationErrorField {
	validationErrors := make([]validationErrorField, 0)

	for index, attachment := range attachments {
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

func newCreatePostAttachments(attachments []createPostAttachmentRequest) []usecase.CreatePostAttachmentCommand {
	result := make([]usecase.CreatePostAttachmentCommand, 0, len(attachments))
	for _, attachment := range attachments {
		result = append(result, usecase.CreatePostAttachmentCommand{
			Kind:    attachment.Kind,
			FileURL: attachment.FileURL,
		})
	}

	return result
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
