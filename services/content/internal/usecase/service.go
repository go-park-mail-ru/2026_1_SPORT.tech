package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

const (
	svgContentType = "image/svg+xml"
)

type Service struct {
	posts           PostRepository
	money           MonetizationRepository
	engagement      EngagementRepository
	notifications   NotificationRepository
	postMedia       PostMediaStorage
	paymentProvider PaymentProvider
}

func NewService(repositories Repositories, postMediaStorage PostMediaStorage, paymentProviders ...PaymentProvider) *Service {
	var paymentProvider PaymentProvider
	if len(paymentProviders) > 0 {
		paymentProvider = paymentProviders[0]
	}

	return &Service{
		posts:           repositories.Posts,
		money:           repositories.Money,
		engagement:      repositories.Engagement,
		notifications:   repositories.Notifications,
		postMedia:       postMediaStorage,
		paymentProvider: paymentProvider,
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

func (service *Service) ListSubscriptionTiers(ctx context.Context, query ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error) {
	if query.TrainerUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	return service.money.ListSubscriptionTiers(ctx, query.TrainerUserID)
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

	return service.money.CreateSubscriptionTier(ctx, tier)
}

func (service *Service) UpdateSubscriptionTier(ctx context.Context, command UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error) {
	if err := validateUpdateSubscriptionTierCommand(command); err != nil {
		return domain.SubscriptionTier{}, err
	}

	tier, err := service.money.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
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

	return service.money.UpdateSubscriptionTier(ctx, tier)
}

func (service *Service) DeleteSubscriptionTier(ctx context.Context, command DeleteSubscriptionTierCommand) error {
	if err := validateSubscriptionTierIDCommand(command.TrainerUserID, command.TierID); err != nil {
		return err
	}

	return service.money.DeleteSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
}

func (service *Service) SubscribeToTrainer(ctx context.Context, command SubscribeToTrainerCommand) (domain.Subscription, error) {
	return domain.Subscription{}, ErrSubscriptionPaymentRequired
}

func (service *Service) createPaidSubscription(ctx context.Context, command SubscribeToTrainerCommand) (domain.Subscription, error) {
	if err := validateSubscribeToTrainerCommand(command); err != nil {
		return domain.Subscription{}, err
	}

	tier, err := service.money.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
	if err != nil {
		return domain.Subscription{}, err
	}

	subscription, err := service.money.SubscribeToTrainer(ctx, domain.Subscription{
		ClientUserID:  command.ClientUserID,
		TrainerUserID: command.TrainerUserID,
		TierID:        tier.TierID,
		ExpiresAt:     time.Now().UTC().AddDate(0, 1, 0),
	})
	if err != nil {
		return domain.Subscription{}, err
	}

	if err := service.createSubscriptionNotifications(ctx, subscription); err != nil {
		return domain.Subscription{}, err
	}

	return subscription, nil
}

func (service *Service) ListMySubscriptions(ctx context.Context, query ListMySubscriptionsQuery) ([]domain.Subscription, error) {
	if query.ClientUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	return service.money.ListSubscriptions(ctx, query.ClientUserID)
}

func (service *Service) ListTrainerSubscribers(ctx context.Context, query ListTrainerSubscribersQuery) ([]domain.Subscription, error) {
	if query.TrainerUserID <= 0 {
		return nil, ErrInvalidUserID
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return service.money.ListTrainerSubscribers(ctx, query.TrainerUserID, limit, offset)
}

func (service *Service) UpdateSubscription(ctx context.Context, command UpdateSubscriptionCommand) (domain.Subscription, error) {
	return domain.Subscription{}, ErrSubscriptionPaymentRequired
}

func (service *Service) updatePaidSubscription(ctx context.Context, command UpdateSubscriptionCommand) (domain.Subscription, error) {
	if err := validateUpdateSubscriptionCommand(command); err != nil {
		return domain.Subscription{}, err
	}

	return service.money.UpdateSubscription(ctx, domain.Subscription{
		SubscriptionID: command.SubscriptionID,
		ClientUserID:   command.ClientUserID,
		TierID:         command.TierID,
	})
}

func (service *Service) CancelSubscription(ctx context.Context, command CancelSubscriptionCommand) error {
	if err := validateSubscriptionIDCommand(command.ClientUserID, command.SubscriptionID); err != nil {
		return err
	}

	return service.money.CancelSubscription(ctx, command.ClientUserID, command.SubscriptionID)
}

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
		if err := service.createNotification(ctx, domain.Notification{
			UserID:      post.AuthorUserID,
			Type:        domain.NotificationTypeLike,
			ActorUserID: command.UserID,
			Title:       "Новый лайк",
			Body:        "Пользователь оценил ваш пост",
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

func (service *Service) DonateToProfile(ctx context.Context, command DonateToProfileCommand) (domain.Donation, error) {
	command.Currency = normalizeCurrency(command.Currency)
	command.Message = normalizeOptionalText(command.Message)
	if err := validateDonateToProfileCommand(command); err != nil {
		return domain.Donation{}, err
	}

	donation, err := service.money.CreateDonation(ctx, domain.Donation{
		SenderUserID:    command.SenderUserID,
		RecipientUserID: command.RecipientUserID,
		AmountValue:     command.AmountValue,
		Currency:        command.Currency,
		Message:         command.Message,
	})
	if err != nil {
		return domain.Donation{}, err
	}

	if err := service.createNotification(ctx, domain.Notification{
		UserID:      donation.RecipientUserID,
		Type:        domain.NotificationTypeDonation,
		ActorUserID: donation.SenderUserID,
		Title:       "Новый донат",
		Body:        "Пользователь отправил вам донат",
		DonationID:  &donation.DonationID,
	}); err != nil {
		return domain.Donation{}, err
	}

	return donation, nil
}

func (service *Service) CreateDonationPayment(ctx context.Context, command CreateDonationPaymentCommand) (domain.DonationPayment, error) {
	donationCommand := DonateToProfileCommand{
		SenderUserID:    command.SenderUserID,
		RecipientUserID: command.RecipientUserID,
		AmountValue:     command.AmountValue,
		Currency:        normalizeCurrency(command.Currency),
		Message:         normalizeOptionalText(command.Message),
	}
	if err := validateDonateToProfileCommand(donationCommand); err != nil {
		return domain.DonationPayment{}, err
	}

	confirmationToken, err := randomToken("confirm")
	if err != nil {
		return domain.DonationPayment{}, err
	}
	if service.paymentProvider == nil {
		return domain.DonationPayment{}, ErrPaymentProviderUnavailable
	}

	payment := domain.DonationPayment{
		Provider:          service.paymentProvider.ProviderName(),
		Status:            domain.PaymentStatusPending,
		SenderUserID:      donationCommand.SenderUserID,
		RecipientUserID:   donationCommand.RecipientUserID,
		AmountValue:       donationCommand.AmountValue,
		Currency:          donationCommand.Currency,
		Message:           donationCommand.Message,
		ConfirmationToken: confirmationToken,
	}
	created, err := service.money.CreateDonationPayment(ctx, payment)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	providerPayment, err := service.paymentProvider.CreatePayment(ctx, PaymentProviderCreateRequest{
		AmountValue:    donationCommand.AmountValue,
		Currency:       donationCommand.Currency,
		Description:    fmt.Sprintf("Donation payment #%d", created.PaymentID),
		IdempotenceKey: paymentIdempotenceKey("donation", created),
		ReturnURL:      normalizeOptionalURL(command.ReturnURL),
		CancelURL:      normalizeOptionalURL(command.CancelURL),
	})
	if err != nil {
		return domain.DonationPayment{}, fmt.Errorf("%w: %v", ErrPaymentProviderUnavailable, err)
	}

	return service.money.UpdateDonationPaymentProvider(ctx, created.PaymentID, providerPayment.ProviderPaymentID, providerPayment.ConfirmationURL)
}

func (service *Service) CreateSubscriptionPayment(ctx context.Context, command CreateSubscriptionPaymentCommand) (domain.DonationPayment, error) {
	if command.ClientUserID <= 0 || command.TrainerUserID <= 0 {
		return domain.DonationPayment{}, ErrInvalidUserID
	}
	if command.ClientUserID == command.TrainerUserID {
		return domain.DonationPayment{}, ErrInvalidSubscriptionTarget
	}
	if command.TierID <= 0 {
		return domain.DonationPayment{}, ErrInvalidSubscriptionTierID
	}

	tier, err := service.money.GetSubscriptionTier(ctx, command.TrainerUserID, command.TierID)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	// Бесплатный уровень оформляем сразу, без платёжного провайдера
	// (провайдер не умеет создавать платёж на 0).
	if tier.Price == 0 {
		return service.createFreeSubscription(ctx, command, tier.TierID)
	}

	if service.paymentProvider == nil {
		return domain.DonationPayment{}, ErrPaymentProviderUnavailable
	}

	confirmationToken, err := randomToken("confirm")
	if err != nil {
		return domain.DonationPayment{}, err
	}
	tierID := tier.TierID
	payment := domain.DonationPayment{
		Provider:          service.paymentProvider.ProviderName(),
		Status:            domain.PaymentStatusPending,
		SenderUserID:      command.ClientUserID,
		RecipientUserID:   command.TrainerUserID,
		AmountValue:       tier.Price,
		Currency:          "RUB",
		ConfirmationToken: confirmationToken,
		TierID:            &tierID,
	}
	created, err := service.money.CreateDonationPayment(ctx, payment)
	if err != nil {
		return domain.DonationPayment{}, err
	}

	providerPayment, err := service.paymentProvider.CreatePayment(ctx, PaymentProviderCreateRequest{
		AmountValue:    tier.Price,
		Currency:       payment.Currency,
		Description:    fmt.Sprintf("Subscription payment #%d", created.PaymentID),
		IdempotenceKey: paymentIdempotenceKey("subscription", created),
		ReturnURL:      normalizeOptionalURL(command.ReturnURL),
		CancelURL:      normalizeOptionalURL(command.CancelURL),
	})
	if err != nil {
		return domain.DonationPayment{}, fmt.Errorf("%w: %v", ErrPaymentProviderUnavailable, err)
	}

	return service.money.UpdateDonationPaymentProvider(ctx, created.PaymentID, providerPayment.ProviderPaymentID, providerPayment.ConfirmationURL)
}

func (service *Service) createFreeSubscription(ctx context.Context, command CreateSubscriptionPaymentCommand, tierID int64) (domain.DonationPayment, error) {
	subscription, err := service.createPaidSubscription(ctx, SubscribeToTrainerCommand{
		ClientUserID:  command.ClientUserID,
		TrainerUserID: command.TrainerUserID,
		TierID:        command.TierID,
	})
	if err != nil {
		return domain.DonationPayment{}, err
	}

	return domain.DonationPayment{
		Status:          domain.PaymentStatusConfirmed,
		SenderUserID:    command.ClientUserID,
		RecipientUserID: command.TrainerUserID,
		AmountValue:     0,
		Currency:        "RUB",
		TierID:          &tierID,
		Subscription:    &subscription,
	}, nil
}

func (service *Service) ConfirmDonationPayment(ctx context.Context, command ConfirmDonationPaymentCommand) (domain.DonationPayment, error) {
	if command.SenderUserID <= 0 {
		return domain.DonationPayment{}, ErrInvalidUserID
	}
	if command.PaymentID <= 0 {
		return domain.DonationPayment{}, ErrInvalidPaymentID
	}
	if normalizeRequiredText(command.ConfirmationToken) == "" {
		return domain.DonationPayment{}, ErrInvalidPaymentConfirmationToken
	}

	payment, err := service.money.GetDonationPayment(ctx, command.SenderUserID, command.PaymentID)
	if err != nil {
		return domain.DonationPayment{}, err
	}
	if payment.ConfirmationToken != normalizeRequiredText(command.ConfirmationToken) {
		return domain.DonationPayment{}, domain.ErrPaymentTokenMismatch
	}
	wasPending := payment.Status != domain.PaymentStatusConfirmed
	if payment.Status != domain.PaymentStatusConfirmed {
		if service.paymentProvider == nil {
			return domain.DonationPayment{}, ErrPaymentProviderUnavailable
		}
		providerCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()

		providerPayment, err := service.paymentProvider.GetPayment(providerCtx, payment.ProviderPaymentID)
		if err != nil {
			return domain.DonationPayment{}, fmt.Errorf("%w: %v", ErrPaymentProviderUnavailable, err)
		}
		if providerPayment.Status != "succeeded" {
			return domain.DonationPayment{}, ErrPaymentNotSucceeded
		}
	}

	payment, err = service.money.ConfirmDonationPayment(ctx, command.SenderUserID, command.PaymentID, normalizeRequiredText(command.ConfirmationToken))
	if err != nil {
		return domain.DonationPayment{}, err
	}
	if wasPending && payment.Donation != nil {
		if err := service.createNotification(ctx, domain.Notification{
			UserID:      payment.Donation.RecipientUserID,
			Type:        domain.NotificationTypeDonation,
			ActorUserID: payment.Donation.SenderUserID,
			Title:       "Новый донат",
			Body:        "Пользователь отправил вам донат",
			DonationID:  &payment.Donation.DonationID,
		}); err != nil {
			return domain.DonationPayment{}, err
		}
	}
	if wasPending && payment.Subscription != nil {
		if err := service.createSubscriptionNotifications(ctx, *payment.Subscription); err != nil {
			return domain.DonationPayment{}, err
		}
	}

	return payment, nil
}

func (service *Service) GetBalance(ctx context.Context, query GetBalanceQuery) (domain.Balance, error) {
	query.Currency = normalizeCurrency(query.Currency)
	if err := validateGetBalanceQuery(query); err != nil {
		return domain.Balance{}, err
	}

	return service.money.GetBalance(ctx, query.TrainerUserID, query.Currency)
}

func (service *Service) GetTrainerStatistics(ctx context.Context, query GetTrainerStatisticsQuery) (domain.TrainerStatistics, error) {
	query.Currency = normalizeCurrency(query.Currency)
	if err := validateGetTrainerStatisticsQuery(query); err != nil {
		return domain.TrainerStatistics{}, err
	}

	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	return service.money.GetTrainerStatistics(ctx, query.TrainerUserID, query.Currency, monthStart)
}

func (service *Service) ListNotifications(ctx context.Context, query ListNotificationsQuery) ([]domain.Notification, error) {
	if query.UserID <= 0 {
		return nil, ErrInvalidUserID
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return service.notifications.ListNotifications(ctx, query.UserID, limit, offset)
}

func (service *Service) MarkNotificationRead(ctx context.Context, command MarkNotificationReadCommand) (domain.Notification, error) {
	if command.UserID <= 0 {
		return domain.Notification{}, ErrInvalidUserID
	}
	if command.NotificationID <= 0 {
		return domain.Notification{}, ErrInvalidNotificationID
	}

	return service.notifications.MarkNotificationRead(ctx, command.UserID, command.NotificationID)
}

func (service *Service) notifySubscribersAboutPost(ctx context.Context, post domain.Post) error {
	if service.notifications == nil {
		return nil
	}

	for offset := int32(0); ; offset += maxPageLimit {
		subscribers, err := service.money.ListTrainerSubscribers(ctx, post.AuthorUserID, maxPageLimit, offset)
		if err != nil {
			return err
		}
		if len(subscribers) == 0 {
			return nil
		}

		for _, subscriber := range subscribers {
			if post.RequiredSubscriptionLevel != nil && subscriber.TierID < int64(*post.RequiredSubscriptionLevel) {
				continue
			}
			if err := service.createNotification(ctx, domain.Notification{
				UserID:      subscriber.ClientUserID,
				Type:        domain.NotificationTypePost,
				ActorUserID: post.AuthorUserID,
				Title:       "Новый пост",
				Body:        "Автор, на которого вы подписаны, опубликовал новый материал",
				PostID:      &post.PostID,
			}); err != nil {
				return err
			}
		}

		if len(subscribers) < maxPageLimit {
			return nil
		}
	}
}

func (service *Service) createSubscriptionNotifications(ctx context.Context, subscription domain.Subscription) error {
	isNewSubscription := subscription.CreatedAt.Equal(subscription.UpdatedAt)
	tierInfo := subscription.TierName
	if subscription.Price > 0 {
		tierInfo = fmt.Sprintf("%s · %d ₽", tierInfo, subscription.Price)
	}
	trainerTitle := "Новая подписка"
	trainerBody := fmt.Sprintf("Оформлена подписка «%s»", tierInfo)
	clientTitle := "Подписка оформлена"
	clientBody := fmt.Sprintf("Доступ к материалам открыт на месяц · %s", tierInfo)
	if !isNewSubscription {
		trainerTitle = "Подписка обновлена"
		trainerBody = fmt.Sprintf("Обновлён тариф на «%s»", tierInfo)
		clientTitle = "Тариф обновлён"
		clientBody = fmt.Sprintf("Новый тариф уже активен · %s", tierInfo)
	}

	if err := service.createNotification(ctx, domain.Notification{
		UserID:         subscription.TrainerUserID,
		Type:           domain.NotificationTypeSubscription,
		ActorUserID:    subscription.ClientUserID,
		Title:          trainerTitle,
		Body:           trainerBody,
		SubscriptionID: &subscription.SubscriptionID,
	}); err != nil {
		return err
	}

	return service.createNotification(ctx, domain.Notification{
		UserID:         subscription.ClientUserID,
		Type:           domain.NotificationTypeSubscription,
		ActorUserID:    subscription.TrainerUserID,
		Title:          clientTitle,
		Body:           clientBody,
		SubscriptionID: &subscription.SubscriptionID,
	})
}

func (service *Service) createNotification(ctx context.Context, notification domain.Notification) error {
	if service.notifications == nil {
		return nil
	}

	notification.Title = normalizeRequiredText(notification.Title)
	notification.Body = normalizeRequiredText(notification.Body)
	if err := validateNotification(notification); err != nil {
		return err
	}

	_, err := service.notifications.CreateNotification(ctx, notification)
	return err
}

func randomToken(prefix string) (string, error) {
	var data [18]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return prefix + "_" + base64.RawURLEncoding.EncodeToString(data[:]), nil
}

func paymentIdempotenceKey(kind string, payment domain.DonationPayment) string {
	return fmt.Sprintf("content-%s-payment-%d-%s", kind, payment.PaymentID, payment.ConfirmationToken)
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

func normalizeOptionalURL(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
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

	_, err := service.money.GetSubscriptionTier(ctx, trainerUserID, int64(*level))
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

	viewerSubscriptionLevel, err := service.money.GetActiveSubscriptionLevel(ctx, viewerUserID, authorUserID)
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
