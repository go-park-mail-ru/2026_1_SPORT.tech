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

func (server *Server) SearchPosts(ctx context.Context, request *gatewayv1.SearchPostsRequest) (*gatewayv1.SearchPostsResponse, error) {
	principal, err := server.optionalSession(ctx)
	if err != nil {
		return nil, err
	}

	var viewerUserID int64
	if principal != nil && principal.User != nil {
		viewerUserID = principal.User.GetUserId()
	}
	viewerSubscriptionLevel, err := subscriptionLevelFromContext(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.SearchPosts(
		forwardContext(ctx),
		mappers.SearchPostsRequestToContent(viewerUserID, viewerSubscriptionLevel, request),
	)
	if err != nil {
		return nil, err
	}

	return mappers.SearchPostsResponseFromContent(response)
}

func (server *Server) CreatePost(ctx context.Context, request *gatewayv1.CreatePostRequest) (*gatewayv1.PostResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	if err := mappers.RequireTrainerRole(principal.User); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.contentClient.CreatePost(
		forwardContext(ctx),
		mappers.CreatePostRequestToContent(userID, request),
	)
	if err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 201); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return mappers.PostResponseFromContent(response)
}

func (server *Server) UploadPostMedia(ctx context.Context, request *gatewayv1.UploadPostMediaRequest) (*gatewayv1.PostMediaUploadResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	if err := mappers.RequireTrainerRole(principal.User); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.contentClient.UploadPostMedia(
		forwardContext(ctx),
		mappers.UploadPostMediaRequestToContent(userID, request),
	)
	if err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 201); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return mappers.PostMediaUploadResponseFromContent(response)
}

func (server *Server) GetPost(ctx context.Context, request *gatewayv1.GetPostRequest) (*gatewayv1.PostResponse, error) {
	principal, err := server.optionalSession(ctx)
	if err != nil {
		return nil, err
	}

	var viewerUserID int64
	if principal != nil && principal.User != nil {
		viewerUserID = principal.User.GetUserId()
	}
	viewerSubscriptionLevel, err := subscriptionLevelFromContext(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.GetPost(
		forwardContext(ctx),
		&contentv1.GetPostRequest{
			PostId:                  int64(request.GetPostId()),
			ViewerUserId:            viewerUserID,
			ViewerSubscriptionLevel: viewerSubscriptionLevel,
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.PostResponseFromContent(response)
}

func (server *Server) UpdatePost(ctx context.Context, request *gatewayv1.UpdatePostRequest) (*gatewayv1.PostResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	if err := mappers.RequireTrainerRole(principal.User); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.contentClient.UpdatePost(
		forwardContext(ctx),
		mappers.UpdatePostRequestToContent(userID, request),
	)
	if err != nil {
		return nil, err
	}

	return mappers.PostResponseFromContent(response)
}

func (server *Server) DeletePost(ctx context.Context, request *gatewayv1.DeletePostRequest) (*emptypb.Empty, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}
	if err := mappers.RequireTrainerRole(principal.User); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if _, err := server.contentClient.DeletePost(
		forwardContext(ctx),
		&contentv1.DeletePostRequest{
			PostId:       int64(request.GetPostId()),
			AuthorUserId: userID,
		},
	); err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 204); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) LikePost(ctx context.Context, request *gatewayv1.PostLikeRequest) (*gatewayv1.PostLikeResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	viewerSubscriptionLevel, err := subscriptionLevelFromContext(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.LikePost(
		forwardContext(ctx),
		&contentv1.LikePostRequest{
			PostId:                  int64(request.GetPostId()),
			UserId:                  userID,
			ViewerSubscriptionLevel: viewerSubscriptionLevel,
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.PostLikeResponseFromContent(response)
}

func (server *Server) UnlikePost(ctx context.Context, request *gatewayv1.PostLikeRequest) (*gatewayv1.PostLikeResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	viewerSubscriptionLevel, err := subscriptionLevelFromContext(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.contentClient.UnlikePost(
		forwardContext(ctx),
		&contentv1.UnlikePostRequest{
			PostId:                  int64(request.GetPostId()),
			UserId:                  userID,
			ViewerSubscriptionLevel: viewerSubscriptionLevel,
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.PostLikeResponseFromContent(response)
}
