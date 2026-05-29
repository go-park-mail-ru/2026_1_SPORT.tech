package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

const (
	svgContentType = "image/svg+xml"
)

type Service struct {
	posts                   PostRepository
	money                   MonetizationRepository
	engagement              EngagementRepository
	notifications           NotificationRepository
	notificationPreferences NotificationPreferencesRepository
	chat                    ChatRepository
	meeting                 MeetingRepository
	postMedia               PostMediaStorage
	paymentProvider         PaymentProvider
}

func NewService(repositories Repositories, postMediaStorage PostMediaStorage, paymentProviders ...PaymentProvider) *Service {
	var paymentProvider PaymentProvider
	if len(paymentProviders) > 0 {
		paymentProvider = paymentProviders[0]
	}

	return &Service{
		posts:                   repositories.Posts,
		money:                   repositories.Money,
		engagement:              repositories.Engagement,
		notifications:           repositories.Notifications,
		notificationPreferences: repositories.NotificationPreferences,
		chat:                    repositories.Chat,
		meeting:                 repositories.Meeting,
		postMedia:               postMediaStorage,
		paymentProvider:         paymentProvider,
	}
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

func (service *Service) ensureRequiredSubscriptionTier(ctx context.Context, trainerUserID int64, level *int32) error {
	if level == nil {
		return nil
	}

	_, err := service.money.GetSubscriptionTier(ctx, trainerUserID, int64(*level))
	return err
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

func randomToken(prefix string) (string, error) {
	var data [18]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	return prefix + "_" + base64.RawURLEncoding.EncodeToString(data[:]), nil
}

func paymentIdempotenceKey(kind string, confirmationToken string) string {
	return fmt.Sprintf("content-%s-payment-%s", kind, confirmationToken)
}
