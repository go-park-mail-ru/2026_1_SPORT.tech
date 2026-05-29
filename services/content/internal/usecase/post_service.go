package usecase

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func (service *Service) ListAuthorPosts(ctx context.Context, query ListAuthorPostsQuery) ([]domain.PostSummary, error) {
	if err := validateListAuthorPostsQuery(query); err != nil {
		return nil, err
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	posts, err := service.posts.ListAuthorPosts(ctx, query.AuthorUserID, query.ViewerUserID, limit, offset)
	if err != nil {
		return nil, err
	}

	for index := range posts {
		canView, err := service.canViewPost(
			ctx,
			posts[index].RequiredSubscriptionLevel,
			posts[index].AuthorUserID,
			query.ViewerUserID,
		)
		if err != nil {
			return nil, err
		}

		posts[index].CanView = canView
	}

	return posts, nil
}

func (service *Service) SearchPosts(ctx context.Context, query SearchPostsQuery) ([]domain.PostSummary, error) {
	if err := validateSearchPostsQuery(query); err != nil {
		return nil, err
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}
	query.Query = normalizeRequiredText(query.Query)
	query.Limit = limit
	query.Offset = offset

	posts, err := service.posts.SearchPosts(ctx, query)
	if err != nil {
		return nil, err
	}

	for index := range posts {
		canView, err := service.canViewPost(
			ctx,
			posts[index].RequiredSubscriptionLevel,
			posts[index].AuthorUserID,
			query.ViewerUserID,
		)
		if err != nil {
			return nil, err
		}

		posts[index].CanView = canView
	}

	return posts, nil
}

func (service *Service) CreatePost(ctx context.Context, command CreatePostCommand) (domain.Post, error) {
	post, err := buildPost(command)
	if err != nil {
		return domain.Post{}, err
	}
	if err := service.ensureRequiredSubscriptionTier(ctx, post.AuthorUserID, post.RequiredSubscriptionLevel); err != nil {
		return domain.Post{}, err
	}

	postID, err := service.posts.CreatePost(ctx, post)
	if err != nil {
		return domain.Post{}, err
	}

	created, err := service.posts.GetPost(ctx, postID, command.AuthorUserID)
	if err != nil {
		return domain.Post{}, err
	}
	if err := service.notifySubscribersAboutPost(ctx, created); err != nil {
		return domain.Post{}, err
	}

	return created, nil
}

func (service *Service) UploadPostMedia(ctx context.Context, command UploadPostMediaCommand) (domain.PostMedia, error) {
	if err := validateUploadPostMediaCommand(command); err != nil {
		return domain.PostMedia{}, err
	}

	fileName := normalizeRequiredText(command.FileName)
	contentType := strings.ToLower(normalizeRequiredText(command.ContentType))
	kind, ok := postMediaKind(contentType)
	if !ok {
		return domain.PostMedia{}, ErrPostMediaContentTypeUnsupported
	}

	if service.postMedia == nil {
		return domain.PostMedia{}, ErrPostMediaStorageUnavailable
	}

	fileURL, err := service.postMedia.UploadPostMedia(
		ctx,
		command.AuthorUserID,
		fileName,
		contentType,
		bytes.NewReader(command.Content),
		int64(len(command.Content)),
	)
	if err != nil {
		return domain.PostMedia{}, fmt.Errorf("%w: %v", ErrPostMediaStorageUnavailable, err)
	}

	return domain.PostMedia{
		FileURL:     fileURL,
		Kind:        kind,
		ContentType: contentType,
		SizeBytes:   int64(len(command.Content)),
	}, nil
}

func (service *Service) GetPost(ctx context.Context, query GetPostQuery) (domain.Post, error) {
	if err := validatePostQuery(query.PostID, query.ViewerUserID); err != nil {
		return domain.Post{}, err
	}

	post, err := service.posts.GetPost(ctx, query.PostID, query.ViewerUserID)
	if err != nil {
		return domain.Post{}, err
	}
	canView, err := service.canViewPost(ctx, post.RequiredSubscriptionLevel, post.AuthorUserID, query.ViewerUserID)
	if err != nil {
		return domain.Post{}, err
	}
	if !canView {
		return domain.Post{}, domain.ErrPostForbidden
	}

	post.CanView = true

	return post, nil
}

func (service *Service) UpdatePost(ctx context.Context, command UpdatePostCommand) (domain.Post, error) {
	if err := validateUpdatePostCommand(command); err != nil {
		return domain.Post{}, err
	}

	post, err := service.posts.GetPost(ctx, command.PostID, command.AuthorUserID)
	if err != nil {
		return domain.Post{}, err
	}
	if post.AuthorUserID != command.AuthorUserID {
		return domain.Post{}, domain.ErrPostForbidden
	}

	if command.Title != nil {
		post.Title = normalizeRequiredText(*command.Title)
	}
	switch {
	case command.ClearRequiredSubscriptionLevel:
		post.RequiredSubscriptionLevel = nil
	case command.RequiredSubscriptionLevel != nil:
		post.RequiredSubscriptionLevel = normalizeSubscriptionLevel(command.RequiredSubscriptionLevel)
	}
	switch {
	case command.ClearSportTypeID:
		post.SportTypeID = nil
	case command.SportTypeID != nil:
		post.SportTypeID = normalizeSportTypeID(command.SportTypeID)
	}
	if command.ReplaceBlocks {
		post.Blocks = normalizeBlocks(command.Blocks)
	}
	if command.IsPinned != nil {
		post.IsPinned = *command.IsPinned
	}

	if err := validatePost(post); err != nil {
		return domain.Post{}, err
	}
	if err := service.ensureRequiredSubscriptionTier(ctx, post.AuthorUserID, post.RequiredSubscriptionLevel); err != nil {
		return domain.Post{}, err
	}

	if err := service.posts.UpdatePost(ctx, post, command.ReplaceBlocks); err != nil {
		return domain.Post{}, err
	}

	return service.posts.GetPost(ctx, post.PostID, command.AuthorUserID)
}

func (service *Service) DeletePost(ctx context.Context, command DeletePostCommand) error {
	if err := validatePostOwnerCommand(command.PostID, command.AuthorUserID); err != nil {
		return err
	}

	post, err := service.posts.GetPost(ctx, command.PostID, command.AuthorUserID)
	if err != nil {
		return err
	}
	if post.AuthorUserID != command.AuthorUserID {
		return domain.ErrPostForbidden
	}

	return service.posts.DeletePost(ctx, command.PostID, command.AuthorUserID)
}
