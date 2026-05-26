package usecase

import "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"

func validateListAuthorPostsQuery(query ListAuthorPostsQuery) error {
	if query.AuthorUserID <= 0 {
		return ErrInvalidUserID
	}
	if query.ViewerUserID < 0 {
		return ErrInvalidUserID
	}

	return nil
}

func validateSearchPostsQuery(query SearchPostsQuery) error {
	if query.ViewerUserID < 0 {
		return ErrInvalidUserID
	}
	for _, authorUserID := range query.AuthorUserIDs {
		if authorUserID <= 0 {
			return ErrInvalidUserID
		}
	}
	for _, sportTypeID := range query.SportTypeIDs {
		if sportTypeID <= 0 {
			return ErrInvalidSportTypeID
		}
	}
	for _, kind := range query.BlockKinds {
		if !kind.IsValid() {
			return domain.ErrInvalidBlockKind
		}
	}
	if query.MinRequiredSubscriptionLevel != nil && *query.MinRequiredSubscriptionLevel < 0 {
		return ErrInvalidSearchFilter
	}
	if query.MaxRequiredSubscriptionLevel != nil && *query.MaxRequiredSubscriptionLevel < 0 {
		return ErrInvalidSearchFilter
	}
	if query.MinRequiredSubscriptionLevel != nil && query.MaxRequiredSubscriptionLevel != nil &&
		*query.MinRequiredSubscriptionLevel > *query.MaxRequiredSubscriptionLevel {
		return ErrInvalidSearchFilter
	}

	return nil
}

func validateUploadPostMediaCommand(command UploadPostMediaCommand) error {
	if command.AuthorUserID <= 0 {
		return ErrInvalidUserID
	}
	if normalizeRequiredText(command.FileName) == "" {
		return ErrPostMediaFileNameRequired
	}
	if normalizeRequiredText(command.ContentType) == "" {
		return ErrPostMediaContentTypeRequired
	}
	if len(command.Content) == 0 {
		return ErrPostMediaContentRequired
	}
	if len(command.Content) > maxMediaFileSize {
		return ErrPostMediaTooLarge
	}

	return nil
}

func validatePostQuery(postID int64, viewerUserID int64) error {
	if postID <= 0 {
		return ErrInvalidPostID
	}
	if viewerUserID < 0 {
		return ErrInvalidUserID
	}

	return nil
}

func validateUpdatePostCommand(command UpdatePostCommand) error {
	if command.PostID <= 0 {
		return ErrInvalidPostID
	}
	if command.AuthorUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.RequiredSubscriptionLevel != nil && command.ClearRequiredSubscriptionLevel {
		return ErrConflictingSubscriptionLevelUpdate
	}
	if command.SportTypeID != nil && command.ClearSportTypeID {
		return ErrConflictingSportTypeUpdate
	}
	if len(command.Blocks) > 0 && !command.ReplaceBlocks {
		return ErrReplaceBlocksRequired
	}

	return nil
}

func validatePostOwnerCommand(postID int64, authorUserID int64) error {
	if postID <= 0 {
		return ErrInvalidPostID
	}
	if authorUserID <= 0 {
		return ErrInvalidUserID
	}

	return nil
}

func validatePost(post domain.Post) error {
	if post.AuthorUserID <= 0 {
		return ErrInvalidUserID
	}
	if len(post.Title) == 0 || len(post.Title) > 200 {
		return ErrInvalidTitle
	}
	if post.RequiredSubscriptionLevel != nil && *post.RequiredSubscriptionLevel < 1 {
		return ErrInvalidRequiredSubscriptionLevel
	}
	if post.SportTypeID != nil && *post.SportTypeID < 1 {
		return ErrInvalidSportTypeID
	}
	if len(post.Blocks) == 0 {
		return ErrBlocksRequired
	}
	if len(post.Blocks) > maxBlockCount {
		return ErrTooManyBlocks
	}

	for _, block := range post.Blocks {
		if !block.Kind.IsValid() {
			return domain.ErrInvalidBlockKind
		}
		switch block.Kind {
		case domain.BlockKindText:
			if block.TextContent == nil || len(*block.TextContent) == 0 || block.FileURL != nil {
				return domain.ErrInvalidBlockData
			}
		default:
			if block.FileURL == nil || len(*block.FileURL) == 0 || block.TextContent != nil {
				return domain.ErrInvalidBlockData
			}
		}
	}

	return nil
}

func validateLikeCommand(command LikePostCommand) error {
	if command.PostID <= 0 {
		return ErrInvalidPostID
	}
	if command.UserID <= 0 {
		return ErrInvalidUserID
	}

	return nil
}

func validateCreateCommentCommand(command CreateCommentCommand) error {
	if command.PostID <= 0 {
		return ErrInvalidPostID
	}
	if command.AuthorUserID <= 0 {
		return ErrInvalidUserID
	}
	body := normalizeRequiredText(command.Body)
	if len(body) == 0 || len(body) > 2000 {
		return ErrInvalidCommentBody
	}

	return nil
}

func validateListCommentsQuery(query ListCommentsQuery) error {
	if query.PostID <= 0 {
		return ErrInvalidPostID
	}
	if query.ViewerUserID < 0 {
		return ErrInvalidUserID
	}

	return nil
}

func validateListPostLikesQuery(query ListPostLikesQuery) error {
	if query.PostID <= 0 {
		return ErrInvalidPostID
	}
	if query.ViewerUserID < 0 {
		return ErrInvalidUserID
	}

	return nil
}
