package grpc

import (
	"context"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) ListAuthorPosts(ctx context.Context, request *gatewayv1.ListAuthorPostsRequest) (*gatewayv1.ListAuthorPostsResponse, error) {
	response, err := server.contentClient.ListAuthorPosts(forwardContext(ctx), mappers.ListAuthorPostsRequestToContent(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.ListAuthorPostsResponseFromContent(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map list author posts response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) CreatePost(ctx context.Context, request *gatewayv1.CreatePostRequest) (*gatewayv1.PostResponse, error) {
	response, err := server.contentClient.CreatePost(forwardContext(ctx), mappers.CreatePostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.PostResponseFromContent(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map create post response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) GetPost(ctx context.Context, request *gatewayv1.GetPostRequest) (*gatewayv1.PostResponse, error) {
	response, err := server.contentClient.GetPost(forwardContext(ctx), mappers.GetPostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.PostResponseFromContent(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map get post response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) UpdatePost(ctx context.Context, request *gatewayv1.UpdatePostRequest) (*gatewayv1.PostResponse, error) {
	response, err := server.contentClient.UpdatePost(forwardContext(ctx), mappers.UpdatePostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.PostResponseFromContent(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map update post response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) DeletePost(ctx context.Context, request *gatewayv1.DeletePostRequest) (*emptypb.Empty, error) {
	return server.contentClient.DeletePost(forwardContext(ctx), mappers.DeletePostRequestToContent(request))
}

func (server *Server) LikePost(ctx context.Context, request *gatewayv1.LikePostRequest) (*gatewayv1.PostLikeStateResponse, error) {
	response, err := server.contentClient.LikePost(forwardContext(ctx), mappers.LikePostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.PostLikeStateResponseFromContent(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map like post response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) UnlikePost(ctx context.Context, request *gatewayv1.UnlikePostRequest) (*gatewayv1.PostLikeStateResponse, error) {
	response, err := server.contentClient.UnlikePost(forwardContext(ctx), mappers.UnlikePostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.PostLikeStateResponseFromContent(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map unlike post response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) CreateComment(ctx context.Context, request *gatewayv1.CreateCommentRequest) (*gatewayv1.CommentResponse, error) {
	response, err := server.contentClient.CreateComment(forwardContext(ctx), mappers.CreateCommentRequestToContent(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.CommentResponseFromContent(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map create comment response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) ListComments(ctx context.Context, request *gatewayv1.ListCommentsRequest) (*gatewayv1.ListCommentsResponse, error) {
	response, err := server.contentClient.ListComments(forwardContext(ctx), mappers.ListCommentsRequestToContent(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.ListCommentsResponseFromContent(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map list comments response: %v", err)
	}

	return mappedResponse, nil
}
