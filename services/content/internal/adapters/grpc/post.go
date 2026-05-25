package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) ListAuthorPosts(ctx context.Context, request *contentv1.ListAuthorPostsRequest) (*contentv1.ListAuthorPostsResponse, error) {
	posts, err := server.useCases.Posts.ListAuthorPosts(ctx, mappers.ListAuthorPostsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListAuthorPostsResponse(posts), nil
}

func (server *Server) SearchPosts(ctx context.Context, request *contentv1.SearchPostsRequest) (*contentv1.SearchPostsResponse, error) {
	posts, err := server.useCases.Posts.SearchPosts(ctx, mappers.SearchPostsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSearchPostsResponse(posts), nil
}

func (server *Server) CreatePost(ctx context.Context, request *contentv1.CreatePostRequest) (*contentv1.PostResponse, error) {
	post, err := server.useCases.Posts.CreatePost(ctx, mappers.CreatePostRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostResponse(post), nil
}

func (server *Server) UploadPostMedia(ctx context.Context, request *contentv1.UploadPostMediaRequest) (*contentv1.PostMediaResponse, error) {
	media, err := server.useCases.PostMedia.UploadPostMedia(ctx, mappers.UploadPostMediaRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostMediaResponse(media), nil
}

func (server *Server) GetPost(ctx context.Context, request *contentv1.GetPostRequest) (*contentv1.PostResponse, error) {
	post, err := server.useCases.Posts.GetPost(ctx, mappers.GetPostRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostResponse(post), nil
}

func (server *Server) UpdatePost(ctx context.Context, request *contentv1.UpdatePostRequest) (*contentv1.PostResponse, error) {
	post, err := server.useCases.Posts.UpdatePost(ctx, mappers.UpdatePostRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostResponse(post), nil
}

func (server *Server) DeletePost(ctx context.Context, request *contentv1.DeletePostRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Posts.DeletePost(ctx, mappers.DeletePostRequestToCommand(request)); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.Empty(), nil
}

func (server *Server) LikePost(ctx context.Context, request *contentv1.LikePostRequest) (*contentv1.PostLikeStateResponse, error) {
	state, err := server.useCases.Posts.LikePost(ctx, mappers.LikePostRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostLikeStateResponse(state), nil
}

func (server *Server) UnlikePost(ctx context.Context, request *contentv1.UnlikePostRequest) (*contentv1.PostLikeStateResponse, error) {
	state, err := server.useCases.Posts.UnlikePost(ctx, mappers.UnlikePostRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostLikeStateResponse(state), nil
}

func (server *Server) ListPostLikes(ctx context.Context, request *contentv1.ListPostLikesRequest) (*contentv1.ListPostLikesResponse, error) {
	likes, err := server.useCases.Posts.ListPostLikes(ctx, mappers.ListPostLikesRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListPostLikesResponse(likes), nil
}
