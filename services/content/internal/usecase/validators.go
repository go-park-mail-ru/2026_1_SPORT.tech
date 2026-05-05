package usecase

import (
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

const (
	defaultPageLimit      = 20
	maxPageLimit          = 100
	maxBlockCount         = 100
	maxMediaFileSize      = 10 * 1024 * 1024
	maxTierNameLen        = 80
	maxTierDescLen        = 500
	maxDonationMessageLen = 500
	defaultCurrency       = "RUB"
)

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
	if strings.TrimSpace(command.FileName) == "" {
		return ErrPostMediaFileNameRequired
	}
	if strings.TrimSpace(command.ContentType) == "" {
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

func validateSubscriptionTier(tier domain.SubscriptionTier) error {
	if tier.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if tier.TierID < 0 {
		return ErrInvalidSubscriptionTierID
	}
	if len(tier.Name) == 0 || len(tier.Name) > maxTierNameLen {
		return ErrInvalidSubscriptionTierName
	}
	if tier.Price < 0 {
		return ErrInvalidSubscriptionTierPrice
	}
	if tier.Description != nil && len(*tier.Description) > maxTierDescLen {
		return ErrInvalidSubscriptionTierDescription
	}

	return nil
}

func validateUpdateSubscriptionTierCommand(command UpdateSubscriptionTierCommand) error {
	if command.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.TierID <= 0 {
		return ErrInvalidSubscriptionTierID
	}
	if command.Description != nil && command.ClearDescription {
		return ErrConflictingTierDescriptionUpdate
	}

	return nil
}

func validateSubscriptionTierIDCommand(trainerUserID int64, tierID int64) error {
	if trainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if tierID <= 0 {
		return ErrInvalidSubscriptionTierID
	}

	return nil
}

func validateSubscribeToTrainerCommand(command SubscribeToTrainerCommand) error {
	if command.ClientUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.ClientUserID == command.TrainerUserID {
		return ErrInvalidSubscriptionTarget
	}
	if command.TierID <= 0 {
		return ErrInvalidSubscriptionTierID
	}

	return nil
}

func validateUpdateSubscriptionCommand(command UpdateSubscriptionCommand) error {
	if command.ClientUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.SubscriptionID <= 0 {
		return ErrInvalidSubscriptionID
	}
	if command.TierID <= 0 {
		return ErrInvalidSubscriptionTierID
	}

	return nil
}

func validateSubscriptionIDCommand(clientUserID int64, subscriptionID int64) error {
	if clientUserID <= 0 {
		return ErrInvalidUserID
	}
	if subscriptionID <= 0 {
		return ErrInvalidSubscriptionID
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

func validateDonateToProfileCommand(command DonateToProfileCommand) error {
	if command.SenderUserID <= 0 || command.RecipientUserID <= 0 {
		return ErrInvalidUserID
	}
	if command.SenderUserID == command.RecipientUserID {
		return ErrInvalidDonationTarget
	}
	if command.AmountValue <= 0 {
		return ErrInvalidDonationAmount
	}
	if normalizeCurrency(command.Currency) != defaultCurrency {
		return ErrInvalidDonationCurrency
	}
	if command.Message != nil && len(normalizeRequiredText(*command.Message)) > maxDonationMessageLen {
		return ErrInvalidDonationMessage
	}

	return nil
}

func validateGetBalanceQuery(query GetBalanceQuery) error {
	if query.TrainerUserID <= 0 {
		return ErrInvalidUserID
	}
	if normalizeCurrency(query.Currency) != defaultCurrency {
		return ErrInvalidDonationCurrency
	}

	return nil
}

func normalizePage(limit int32, offset int32) (int32, int32, error) {
	if limit < 0 || limit > maxPageLimit {
		return 0, 0, ErrInvalidLimit
	}
	if offset < 0 {
		return 0, 0, ErrInvalidOffset
	}
	if limit == 0 {
		limit = defaultPageLimit
	}

	return limit, offset, nil
}
