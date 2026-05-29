package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func (service *Service) LikePost(ctx context.Context, command LikePostCommand) (domain.PostLikeState, error) {
	if err := validateLikeCommand(command); err != nil {
		return domain.PostLikeState{}, err
	}

	post, err := service.posts.GetPost(ctx, command.PostID, command.UserID)
	if err != nil {
		return domain.PostLikeState{}, err
	}
	canView, err := service.canViewPost(ctx, post.RequiredSubscriptionLevel, post.AuthorUserID, command.UserID)
	if err != nil {
		return domain.PostLikeState{}, err
	}
	if !canView {
		return domain.PostLikeState{}, domain.ErrPostForbidden
	}

	wasCreated, err := service.engagement.UpsertLike(ctx, command.PostID, command.UserID)
	if err != nil {
		return domain.PostLikeState{}, err
	}
	if wasCreated && post.AuthorUserID != command.UserID {
		body := fmt.Sprintf("Пользователь оценил пост «%s»", post.Title)
		if err := service.createNotification(ctx, domain.Notification{
			UserID:      post.AuthorUserID,
			Type:        domain.NotificationTypeLike,
			ActorUserID: command.UserID,
			Title:       "Новый лайк",
			Body:        body,
			PostID:      &post.PostID,
		}); err != nil {
			return domain.PostLikeState{}, err
		}
	}

	return service.engagement.GetPostLikeState(ctx, command.PostID, command.UserID)
}

func (service *Service) UnlikePost(ctx context.Context, command LikePostCommand) (domain.PostLikeState, error) {
	if err := validateLikeCommand(command); err != nil {
		return domain.PostLikeState{}, err
	}

	post, err := service.posts.GetPost(ctx, command.PostID, command.UserID)
	if err != nil {
		return domain.PostLikeState{}, err
	}
	canView, err := service.canViewPost(ctx, post.RequiredSubscriptionLevel, post.AuthorUserID, command.UserID)
	if err != nil {
		return domain.PostLikeState{}, err
	}
	if !canView {
		return domain.PostLikeState{}, domain.ErrPostForbidden
	}

	if err := service.engagement.DeleteLike(ctx, command.PostID, command.UserID); err != nil {
		return domain.PostLikeState{}, err
	}

	return service.engagement.GetPostLikeState(ctx, command.PostID, command.UserID)
}

func (service *Service) CreateComment(ctx context.Context, command CreateCommentCommand) (domain.Comment, error) {
	if err := validateCreateCommentCommand(command); err != nil {
		return domain.Comment{}, err
	}
	body := normalizeRequiredText(command.Body)

	post, err := service.posts.GetPost(ctx, command.PostID, command.AuthorUserID)
	if err != nil {
		return domain.Comment{}, err
	}
	canView, err := service.canViewPost(ctx, post.RequiredSubscriptionLevel, post.AuthorUserID, command.AuthorUserID)
	if err != nil {
		return domain.Comment{}, err
	}
	if !canView {
		return domain.Comment{}, domain.ErrPostForbidden
	}

	comment, err := service.engagement.CreateComment(ctx, domain.Comment{
		PostID:       command.PostID,
		AuthorUserID: command.AuthorUserID,
		Body:         body,
	})
	if err != nil {
		return domain.Comment{}, err
	}

	if post.AuthorUserID != command.AuthorUserID {
		if err := service.createNotification(ctx, domain.Notification{
			UserID:      post.AuthorUserID,
			Type:        domain.NotificationTypeComment,
			ActorUserID: command.AuthorUserID,
			Title:       "Новый комментарий",
			Body:        "Пользователь написал комментарий к вашему посту",
			PostID:      &comment.PostID,
			CommentID:   &comment.CommentID,
		}); err != nil {
			return domain.Comment{}, err
		}
	}

	return comment, nil
}

func (service *Service) UpdateComment(ctx context.Context, command UpdateCommentCommand) (domain.Comment, error) {
	if command.CommentID <= 0 {
		return domain.Comment{}, domain.ErrCommentNotFound
	}
	if command.AuthorUserID <= 0 {
		return domain.Comment{}, ErrInvalidUserID
	}
	body := normalizeRequiredText(command.Body)
	if len(body) == 0 || len(body) > 2000 {
		return domain.Comment{}, ErrInvalidCommentBody
	}

	comment, err := service.engagement.GetComment(ctx, command.CommentID)
	if err != nil {
		return domain.Comment{}, err
	}
	if comment.AuthorUserID != command.AuthorUserID {
		return domain.Comment{}, domain.ErrCommentForbidden
	}

	now := time.Now().UTC()
	if now.Sub(comment.CreatedAt) > commentEditWindow {
		return domain.Comment{}, domain.ErrCommentEditWindowExpired
	}

	return service.engagement.UpdateComment(ctx, command.CommentID, body, now)
}

func (service *Service) DeleteComment(ctx context.Context, command DeleteCommentCommand) error {
	if command.CommentID <= 0 {
		return domain.ErrCommentNotFound
	}
	if command.AuthorUserID <= 0 {
		return ErrInvalidUserID
	}

	comment, err := service.engagement.GetComment(ctx, command.CommentID)
	if err != nil {
		return err
	}
	if comment.AuthorUserID != command.AuthorUserID {
		return domain.ErrCommentForbidden
	}

	return service.engagement.DeleteComment(ctx, command.CommentID)
}

func (service *Service) ListComments(ctx context.Context, query ListCommentsQuery) ([]domain.Comment, error) {
	if err := validateListCommentsQuery(query); err != nil {
		return nil, err
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	post, err := service.posts.GetPost(ctx, query.PostID, query.ViewerUserID)
	if err != nil {
		return nil, err
	}
	canView, err := service.canViewPost(ctx, post.RequiredSubscriptionLevel, post.AuthorUserID, query.ViewerUserID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, domain.ErrPostForbidden
	}

	return service.engagement.ListComments(ctx, query.PostID, limit, offset)
}

func (service *Service) ListPostLikes(ctx context.Context, query ListPostLikesQuery) ([]domain.PostLike, error) {
	if err := validateListPostLikesQuery(query); err != nil {
		return nil, err
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	post, err := service.posts.GetPost(ctx, query.PostID, query.ViewerUserID)
	if err != nil {
		return nil, err
	}
	canView, err := service.canViewPost(ctx, post.RequiredSubscriptionLevel, post.AuthorUserID, query.ViewerUserID)
	if err != nil {
		return nil, err
	}
	if !canView {
		return nil, domain.ErrPostForbidden
	}

	return service.engagement.ListPostLikes(ctx, query.PostID, limit, offset)
}
