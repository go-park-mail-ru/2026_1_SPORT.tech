package usecase

import (
	"context"
	"strings"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

func (service *Service) SendChatMessage(ctx context.Context, command SendChatMessageCommand) (domain.ChatMessage, error) {
	body := strings.TrimSpace(command.Body)
	if body == "" {
		return domain.ChatMessage{}, domain.ErrInvalidBlockData
	}
	if len(body) > 4000 {
		return domain.ChatMessage{}, domain.ErrInvalidBlockData
	}

	canSend := false

	ok, err := service.chat.HasActiveChatSubscription(ctx, command.SenderUserID, command.ReceiverUserID)
	if err != nil {
		return domain.ChatMessage{}, err
	}
	if ok {
		canSend = true
	}

	if !canSend {
		ok, err = service.chat.IsTrainerOf(ctx, command.SenderUserID, command.ReceiverUserID)
		if err != nil {
			return domain.ChatMessage{}, err
		}
		if ok {
			canSend = true
		}
	}

	if !canSend {
		return domain.ChatMessage{}, domain.ErrChatAccessForbidden
	}

	msg := domain.ChatMessage{
		SenderUserID:   command.SenderUserID,
		ReceiverUserID: command.ReceiverUserID,
		Body:           body,
	}

	return service.chat.SaveChatMessage(ctx, msg)
}

func (service *Service) ListChatMessages(ctx context.Context, query ListChatMessagesQuery) ([]domain.ChatMessage, error) {
	if query.UserID <= 0 || query.OtherUserID <= 0 {
		return nil, domain.ErrChatAccessForbidden
	}

	limit, offset, err := normalizePage(query.Limit, query.Offset)
	if err != nil {
		return nil, err
	}

	return service.chat.ListChatMessages(ctx, query.UserID, query.OtherUserID, limit, offset)
}

func (service *Service) ListChatConversations(ctx context.Context, query ListChatConversationsQuery) ([]domain.ChatConversation, error) {
	if query.UserID <= 0 {
		return nil, domain.ErrChatAccessForbidden
	}

	return service.chat.ListChatConversations(ctx, query.UserID)
}

func (service *Service) MarkChatMessageRead(ctx context.Context, command MarkChatMessageReadCommand) error {
	return service.chat.MarkChatMessageRead(ctx, command.UserID, command.MessageID)
}
