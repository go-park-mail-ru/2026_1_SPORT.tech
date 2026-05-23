package usecase

import "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"

func validateNotification(notification domain.Notification) error {
	if notification.UserID <= 0 || notification.ActorUserID <= 0 {
		return ErrInvalidUserID
	}
	if !notification.Type.IsValid() {
		return ErrInvalidNotificationType
	}
	if normalizeRequiredText(notification.Title) == "" || len(normalizeRequiredText(notification.Title)) > 200 {
		return ErrInvalidNotificationTitle
	}
	if normalizeRequiredText(notification.Body) == "" || len(normalizeRequiredText(notification.Body)) > 1000 {
		return ErrInvalidNotificationBody
	}

	return nil
}
