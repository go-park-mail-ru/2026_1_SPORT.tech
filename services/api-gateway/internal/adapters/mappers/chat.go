package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func ChatMessageFromContent(msg *contentv1.ChatMessage) *gatewayv1.ChatMessage {
	if msg == nil {
		return nil
	}

	return &gatewayv1.ChatMessage{
		MessageId:      msg.GetMessageId(),
		SenderUserId:   msg.GetSenderUserId(),
		ReceiverUserId: msg.GetReceiverUserId(),
		Body:           msg.GetBody(),
		IsRead:         msg.GetIsRead(),
		CreatedAt:      msg.GetCreatedAt(),
	}
}

func ListMessagesResponseFromContent(response *contentv1.ListChatMessagesResponse) *gatewayv1.ListMessagesResponse {
	messages := make([]*gatewayv1.ChatMessage, 0, len(response.GetMessages()))
	for _, msg := range response.GetMessages() {
		messages = append(messages, ChatMessageFromContent(msg))
	}

	return &gatewayv1.ListMessagesResponse{Messages: messages}
}

func ListConversationsResponseFromContent(response *contentv1.ListChatConversationsResponse) *gatewayv1.ListConversationsResponse {
	conversations := make([]*gatewayv1.ChatConversation, 0, len(response.GetConversations()))
	for _, conv := range response.GetConversations() {
		conversations = append(conversations, &gatewayv1.ChatConversation{
			OtherUserId: conv.GetOtherUserId(),
			LastMessage: ChatMessageFromContent(conv.GetLastMessage()),
			UnreadCount: conv.GetUnreadCount(),
		})
	}

	return &gatewayv1.ListConversationsResponse{Conversations: conversations}
}
