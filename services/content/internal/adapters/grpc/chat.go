package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) SendChatMessage(ctx context.Context, request *contentv1.SendChatMessageRequest) (*contentv1.ChatMessage, error) {
	msg, err := server.useCases.Chat.SendChatMessage(ctx, usecase.SendChatMessageCommand{
		SenderUserID:   request.GetSenderUserId(),
		ReceiverUserID: request.GetReceiverUserId(),
		Body:           request.GetBody(),
	})
	if err != nil {
		return nil, server.statusError("SendChatMessage", err)
	}

	return mappers.ChatMessageToProto(msg), nil
}

func (server *Server) ListChatMessages(ctx context.Context, request *contentv1.ListChatMessagesRequest) (*contentv1.ListChatMessagesResponse, error) {
	messages, err := server.useCases.Chat.ListChatMessages(ctx, usecase.ListChatMessagesQuery{
		UserID:      request.GetUserId(),
		OtherUserID: request.GetOtherUserId(),
		Limit:       request.GetLimit(),
		Offset:      request.GetOffset(),
	})
	if err != nil {
		return nil, server.statusError("ListChatMessages", err)
	}

	return mappers.NewListChatMessagesResponse(messages), nil
}

func (server *Server) ListChatConversations(ctx context.Context, request *contentv1.ListChatConversationsRequest) (*contentv1.ListChatConversationsResponse, error) {
	conversations, err := server.useCases.Chat.ListChatConversations(ctx, usecase.ListChatConversationsQuery{
		UserID: request.GetUserId(),
	})
	if err != nil {
		return nil, server.statusError("ListChatConversations", err)
	}

	return mappers.NewListChatConversationsResponse(conversations), nil
}

func (server *Server) MarkChatMessageRead(ctx context.Context, request *contentv1.MarkChatMessageReadRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Chat.MarkChatMessageRead(ctx, usecase.MarkChatMessageReadCommand{
		UserID:    request.GetUserId(),
		MessageID: request.GetMessageId(),
	}); err != nil {
		return nil, server.statusError("MarkChatMessageRead", err)
	}

	return &emptypb.Empty{}, nil
}
