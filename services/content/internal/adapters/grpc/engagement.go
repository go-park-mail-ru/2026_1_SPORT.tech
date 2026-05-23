package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
)

func (server *Server) CreateComment(ctx context.Context, request *contentv1.CreateCommentRequest) (*contentv1.CommentResponse, error) {
	comment, err := server.useCases.Comments.CreateComment(ctx, mappers.CreateCommentRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewCommentResponse(comment), nil
}

func (server *Server) ListComments(ctx context.Context, request *contentv1.ListCommentsRequest) (*contentv1.ListCommentsResponse, error) {
	comments, err := server.useCases.Comments.ListComments(ctx, mappers.ListCommentsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListCommentsResponse(comments), nil
}
