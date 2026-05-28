package usecase

import (
	"context"
	"fmt"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

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

func (service *Service) GetNotificationPreferences(ctx context.Context, query GetNotificationPreferencesQuery) (domain.NotificationPreferences, error) {
	if query.UserID <= 0 {
		return domain.NotificationPreferences{}, ErrInvalidUserID
	}
	if service.notificationPreferences == nil {
		return domain.DefaultNotificationPreferences(), nil
	}

	return service.notificationPreferences.GetNotificationPreferences(ctx, query.UserID)
}

func (service *Service) UpdateNotificationPreferences(ctx context.Context, command UpdateNotificationPreferencesCommand) (domain.NotificationPreferences, error) {
	if command.UserID <= 0 {
		return domain.NotificationPreferences{}, ErrInvalidUserID
	}
	if service.notificationPreferences == nil {
		return domain.NotificationPreferences{}, ErrNotificationPreferencesUnavailable
	}

	if err := service.notificationPreferences.UpsertNotificationPreferences(ctx, command.UserID, command.Preferences); err != nil {
		return domain.NotificationPreferences{}, err
	}

	return service.notificationPreferences.GetNotificationPreferences(ctx, command.UserID)
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

	if service.notificationPreferences != nil {
		preferences, err := service.notificationPreferences.GetNotificationPreferences(ctx, notification.UserID)
		if err != nil {
			return err
		}
		if !preferences.Allows(notification.Type) {
			return nil
		}
	}

	_, err := service.notifications.CreateNotification(ctx, notification)
	return err
}
