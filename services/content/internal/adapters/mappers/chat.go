package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ChatMessageToProto(msg domain.ChatMessage) *contentv1.ChatMessage {
	return &contentv1.ChatMessage{
		MessageId:      msg.MessageID,
		SenderUserId:   msg.SenderUserID,
		ReceiverUserId: msg.ReceiverUserID,
		Body:           msg.Body,
		IsRead:         msg.IsRead,
		CreatedAt:      timestamppb.New(msg.CreatedAt),
	}
}

func NewListChatMessagesResponse(messages []domain.ChatMessage) *contentv1.ListChatMessagesResponse {
	response := &contentv1.ListChatMessagesResponse{
		Messages: make([]*contentv1.ChatMessage, 0, len(messages)),
	}
	for _, msg := range messages {
		response.Messages = append(response.Messages, ChatMessageToProto(msg))
	}

	return response
}

func NewListChatConversationsResponse(conversations []domain.ChatConversation) *contentv1.ListChatConversationsResponse {
	response := &contentv1.ListChatConversationsResponse{
		Conversations: make([]*contentv1.ChatConversation, 0, len(conversations)),
	}
	for _, conv := range conversations {
		response.Conversations = append(response.Conversations, &contentv1.ChatConversation{
			OtherUserId: conv.OtherUserID,
			LastMessage: ChatMessageToProto(conv.LastMessage),
			UnreadCount: conv.UnreadCount,
		})
	}

	return response
}
