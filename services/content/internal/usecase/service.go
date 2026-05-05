package usecase

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

const (
	svgContentType = "image/svg+xml"
)

type Service struct {
	contentRepository ContentRepository
	postMediaStorage  PostMediaStorage
}

func NewService(contentRepository ContentRepository, postMediaStorage PostMediaStorage) *Service {
	return &Service{
		contentRepository: contentRepository,
		postMediaStorage:  postMediaStorage,
	}
}

func (service *Service) ListAuthorPosts(ctx context.Context, query ListAuthorPostsQuery) ([]domain.PostSummary, error) {
	if err := validateListAuthorPostsQuery(query); err != nil {
		return nil, err
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	posts, err := service.contentRepository.ListAuthorPosts(ctx, query.AuthorUserID, query.ViewerUserID, limit, offset)
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

	posts, err := service.contentRepository.SearchPosts(ctx, query)
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

	postID, err := service.contentRepository.CreatePost(ctx, post)
	if err != nil {
		return domain.Post{}, err
	}

	return service.contentRepository.GetPost(ctx, postID, command.AuthorUserID)
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

	if service.postMediaStorage == nil {
		return domain.PostMedia{}, ErrPostMediaStorageUnavailable
	}

	fileURL, err := service.postMediaStorage.UploadPostMedia(
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

	post, err := service.contentRepository.GetPost(ctx, query.PostID, query.ViewerUserID)
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

	post, err := service.contentRepository.GetPost(ctx, command.PostID, command.AuthorUserID)
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

	if err := validatePost(post); err != nil {
		return domain.Post{}, err
	}
	if err := service.ensureRequiredSubscriptionTier(ctx, post.AuthorUserID, post.RequiredSubscriptionLevel); err != nil {
		return domain.Post{}, err
	}

	if err := service.contentRepository.UpdatePost(ctx, post, command.ReplaceBlocks); err != nil {
		return domain.Post{}, err
	}

	return service.contentRepository.GetPost(ctx, post.PostID, command.AuthorUserID)
}

func (service *Service) DeletePost(ctx context.Context, command DeletePostCommand) error {
	if err := validatePostOwnerCommand(command.PostID, command.AuthorUserID); err != nil {
		return err
	}

	post, err := service.contentRepository.GetPost(ctx, command.PostID, command.AuthorUserID)
	if err != nil {
		return err
	}
	if post.AuthorUserID != command.AuthorUserID {
		return domain.ErrPostForbidden
	}

	return service.contentRepository.DeletePost(ctx, command.PostID, command.AuthorUserID)
}

func (service *Service) ListSubscriptionTiers(ctx context.Context, query ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error) {
	if query.TrainerUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	return service.contentRepository.ListSubscriptionTiers(ctx, query.TrainerUserID)
}

func (service *Service) CreateSubscriptionTier(ctx context.Context, command CreateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	tier := domain.SubscriptionTier{
		TrainerUserID: command.TrainerUserID,
		Name:          normalizeRequiredText(command.Name),
		Price:         command.Price,
		Description:   normalizeOptionalText(command.Description),
	}
	if err := validateSubscriptionTier(tier); err != nil {
		return domain.SubscriptionTier{}, err
	}

	return service.contentRepository.CreateSubscriptionTier(ctx, tier)
}

func (service *Service) UpdateSubscriptionTier(ctx context.Context, command UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	if err := validateUpdateSubscriptionTierCommand(command); err != nil {
		return domain.SubscriptionTier{}, err
	}

	tier, err := service.contentRepository.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
	if err != nil {
		return domain.SubscriptionTier{}, err
	}

	if command.Name != nil {
		tier.Name = normalizeRequiredText(*command.Name)
	}
	if command.Price != nil {
		tier.Price = *command.Price
	}
	switch {
	case command.ClearDescription:
		tier.Description = nil
	case command.Description != nil:
		tier.Description = normalizeOptionalText(command.Description)
	}

	if err := validateSubscriptionTier(tier); err != nil {
		return domain.SubscriptionTier{}, err
	}

	return service.contentRepository.UpdateSubscriptionTier(ctx, tier)
}

func (service *Service) DeleteSubscriptionTier(ctx context.Context, command DeleteSubscriptionTierCommand) error {
	if err := validateSubscriptionTierIDCommand(command.TrainerUserID, command.TierID); err != nil {
		return err
	}

	return service.contentRepository.DeleteSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
}

func (service *Service) SubscribeToTrainer(ctx context.Context, command SubscribeToTrainerCommand) (domain.Subscription, error) {
	if err := validateSubscribeToTrainerCommand(command); err != nil {
		return domain.Subscription{}, err
	}

	tier, err := service.contentRepository.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
	if err != nil {
		return domain.Subscription{}, err
	}

	return service.contentRepository.SubscribeToTrainer(ctx, domain.Subscription{
		ClientUserID:  command.ClientUserID,
		TrainerUserID: command.TrainerUserID,
		TierID:        tier.TierID,
		ExpiresAt:     time.Now().UTC().AddDate(0, 1, 0),
	})
}

func (service *Service) ListMySubscriptions(ctx context.Context, query ListMySubscriptionsQuery) ([]domain.Subscription, error) {
	if query.ClientUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	return service.contentRepository.ListSubscriptions(ctx, query.ClientUserID)
}

func (service *Service) UpdateSubscription(ctx context.Context, command UpdateSubscriptionCommand) (domain.Subscription, error) {
	if err := validateUpdateSubscriptionCommand(command); err != nil {
		return domain.Subscription{}, err
	}

	return service.contentRepository.UpdateSubscription(ctx, domain.Subscription{
		SubscriptionID: command.SubscriptionID,
		ClientUserID:   command.ClientUserID,
		TierID:         command.TierID,
	})
}

func (service *Service) CancelSubscription(ctx context.Context, command CancelSubscriptionCommand) error {
	if err := validateSubscriptionIDCommand(command.ClientUserID, command.SubscriptionID); err != nil {
		return err
	}

	return service.contentRepository.CancelSubscription(ctx, command.ClientUserID, command.SubscriptionID)
}

func (service *Service) LikePost(ctx context.Context, command LikePostCommand) (domain.PostLikeState, error) {
	if err := validateLikeCommand(command); err != nil {
		return domain.PostLikeState{}, err
	}

	post, err := service.contentRepository.GetPost(ctx, command.PostID, command.UserID)
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

	if err := service.contentRepository.UpsertLike(ctx, command.PostID, command.UserID); err != nil {
		return domain.PostLikeState{}, err
	}

	return service.contentRepository.GetPostLikeState(ctx, command.PostID, command.UserID)
}

func (service *Service) UnlikePost(ctx context.Context, command LikePostCommand) (domain.PostLikeState, error) {
	if err := validateLikeCommand(command); err != nil {
		return domain.PostLikeState{}, err
	}

	post, err := service.contentRepository.GetPost(ctx, command.PostID, command.UserID)
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

	if err := service.contentRepository.DeleteLike(ctx, command.PostID, command.UserID); err != nil {
		return domain.PostLikeState{}, err
	}

	return service.contentRepository.GetPostLikeState(ctx, command.PostID, command.UserID)
}

func (service *Service) CreateComment(ctx context.Context, command CreateCommentCommand) (domain.Comment, error) {
	if err := validateCreateCommentCommand(command); err != nil {
		return domain.Comment{}, err
	}
	body := normalizeRequiredText(command.Body)

	post, err := service.contentRepository.GetPost(ctx, command.PostID, command.AuthorUserID)
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

	return service.contentRepository.CreateComment(ctx, domain.Comment{
		PostID:       command.PostID,
		AuthorUserID: command.AuthorUserID,
		Body:         body,
	})
}

func (service *Service) ListComments(ctx context.Context, query ListCommentsQuery) ([]domain.Comment, error) {
	if err := validateListCommentsQuery(query); err != nil {
		return nil, err
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	post, err := service.contentRepository.GetPost(ctx, query.PostID, query.ViewerUserID)
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

	return service.contentRepository.ListComments(ctx, query.PostID, limit, offset)
}

func (service *Service) DonateToProfile(ctx context.Context, command DonateToProfileCommand) (domain.Donation, error) {
	command.Currency = normalizeCurrency(command.Currency)
	command.Message = normalizeOptionalText(command.Message)
	if err := validateDonateToProfileCommand(command); err != nil {
		return domain.Donation{}, err
	}

	return service.contentRepository.CreateDonation(ctx, domain.Donation{
		SenderUserID:    command.SenderUserID,
		RecipientUserID: command.RecipientUserID,
		AmountValue:     command.AmountValue,
		Currency:        command.Currency,
		Message:         command.Message,
	})
}

func (service *Service) GetBalance(ctx context.Context, query GetBalanceQuery) (domain.Balance, error) {
	query.Currency = normalizeCurrency(query.Currency)
	if err := validateGetBalanceQuery(query); err != nil {
		return domain.Balance{}, err
	}

	return service.contentRepository.GetBalance(ctx, query.TrainerUserID, query.Currency)
}

func buildPost(command CreatePostCommand) (domain.Post, error) {
	post := domain.Post{
		AuthorUserID:              command.AuthorUserID,
		Title:                     normalizeRequiredText(command.Title),
		RequiredSubscriptionLevel: normalizeSubscriptionLevel(command.RequiredSubscriptionLevel),
		SportTypeID:               normalizeSportTypeID(command.SportTypeID),
		Blocks:                    normalizeBlocks(command.Blocks),
	}

	if err := validatePost(post); err != nil {
		return domain.Post{}, err
	}

	return post, nil
}

func normalizeRequiredText(value string) string {
	return strings.TrimSpace(value)
}

func normalizeOptionalText(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeSubscriptionLevel(value *int32) *int32 {
	if value == nil {
		return nil
	}

	level := *value

	return &level
}

func normalizeSportTypeID(value *int64) *int64 {
	if value == nil {
		return nil
	}

	sportTypeID := *value

	return &sportTypeID
}

func normalizeCurrency(value string) string {
	value = strings.ToUpper(normalizeRequiredText(value))
	if value == "" {
		return defaultCurrency
	}

	return value
}

func (service *Service) ensureRequiredSubscriptionTier(ctx context.Context, trainerUserID int64, level *int32) error {
	if level == nil {
		return nil
	}

	_, err := service.contentRepository.GetSubscriptionTier(ctx, trainerUserID, int64(*level))
	return err
}

func (service *Service) canViewPost(
	ctx context.Context,
	requiredLevel *int32,
	authorUserID int64,
	viewerUserID int64,
) (bool, error) {
	if requiredLevel == nil {
		return true, nil
	}
	if authorUserID == viewerUserID && authorUserID != 0 {
		return true, nil
	}
	if viewerUserID <= 0 {
		return false, nil
	}

	viewerSubscriptionLevel, err := service.contentRepository.GetActiveSubscriptionLevel(ctx, viewerUserID, authorUserID)
	if err != nil {
		return false, err
	}

	return domain.CanViewPost(requiredLevel, authorUserID, viewerUserID, viewerSubscriptionLevel), nil
}

func normalizeBlocks(inputs []PostBlockInput) []domain.PostBlock {
	blocks := make([]domain.PostBlock, 0, len(inputs))
	for index, input := range inputs {
		blocks = append(blocks, domain.PostBlock{
			Position:    int32(index),
			Kind:        input.Kind,
			TextContent: normalizeOptionalText(input.TextContent),
			FileURL:     normalizeOptionalText(input.FileURL),
		})
	}

	return blocks
}

func postMediaKind(contentType string) (domain.BlockKind, bool) {
	if strings.HasPrefix(contentType, "image/") && contentType != svgContentType {
		return domain.BlockKindImage, true
	}

	switch contentType {
	case "video/mp4":
		return domain.BlockKindVideo, true
	case "application/pdf":
		return domain.BlockKindDocument, true
	default:
		return "", false
	}
}
