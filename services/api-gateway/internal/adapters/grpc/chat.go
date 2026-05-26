package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) SendMessage(ctx context.Context, request *gatewayv1.SendMessageRequest) (*gatewayv1.ChatMessage, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.SendChatMessage(
		forwardContext(ctx),
		&contentv1.SendChatMessageRequest{
			SenderUserId:   userID,
			ReceiverUserId: request.GetReceiverUserId(),
			Body:           request.GetBody(),
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.ChatMessageFromContent(response), nil
}

func (server *Server) ListMessages(ctx context.Context, request *gatewayv1.ListMessagesRequest) (*gatewayv1.ListMessagesResponse, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	if request.GetOtherUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid other_user_id")
	}

	response, err := server.contentClient.ListChatMessages(
		forwardContext(ctx),
		&contentv1.ListChatMessagesRequest{
			UserId:      userID,
			OtherUserId: request.GetOtherUserId(),
			Limit:       request.GetLimit(),
			Offset:      request.GetOffset(),
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.ListMessagesResponseFromContent(response), nil
}

func (server *Server) ListConversations(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.ListConversationsResponse, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.ListChatConversations(
		forwardContext(ctx),
		&contentv1.ListChatConversationsRequest{UserId: userID},
	)
	if err != nil {
		return nil, err
	}

	return mappers.ListConversationsResponseFromContent(response), nil
}

func (server *Server) MarkMessageRead(ctx context.Context, request *gatewayv1.MarkMessageReadRequest) (*emptypb.Empty, error) {
	userID, err := server.requireSubscriptionUserID(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := server.contentClient.MarkChatMessageRead(
		forwardContext(ctx),
		&contentv1.MarkChatMessageReadRequest{
			UserId:    userID,
			MessageId: request.GetMessageId(),
		},
	); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
