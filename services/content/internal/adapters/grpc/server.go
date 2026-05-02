package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ContentUseCase interface {
	ListAuthorPosts(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error)
	SearchPosts(ctx context.Context, query usecase.SearchPostsQuery) ([]domain.PostSummary, error)
	CreatePost(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error)
	UploadPostMedia(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error)
	GetPost(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error)
	UpdatePost(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error)
	DeletePost(ctx context.Context, command usecase.DeletePostCommand) error
	LikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	UnlikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	CreateComment(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error)
	ListComments(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error)
}

type Server struct {
	contentv1.UnimplementedContentServiceServer
	contentUseCase ContentUseCase
}

func NewServer(contentUseCase ContentUseCase) *Server {
	return &Server{contentUseCase: contentUseCase}
}

func (server *Server) ListAuthorPosts(ctx context.Context, request *contentv1.ListAuthorPostsRequest) (*contentv1.ListAuthorPostsResponse, error) {
	posts, err := server.contentUseCase.ListAuthorPosts(ctx, mappers.ListAuthorPostsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListAuthorPostsResponse(posts), nil
}

func (server *Server) SearchPosts(ctx context.Context, request *contentv1.SearchPostsRequest) (*contentv1.SearchPostsResponse, error) {
	posts, err := server.contentUseCase.SearchPosts(ctx, mappers.SearchPostsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSearchPostsResponse(posts), nil
}

func (server *Server) CreatePost(ctx context.Context, request *contentv1.CreatePostRequest) (*contentv1.PostResponse, error) {
	post, err := server.contentUseCase.CreatePost(ctx, mappers.CreatePostRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostResponse(post), nil
}

func (server *Server) UploadPostMedia(ctx context.Context, request *contentv1.UploadPostMediaRequest) (*contentv1.PostMediaResponse, error) {
	media, err := server.contentUseCase.UploadPostMedia(ctx, mappers.UploadPostMediaRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostMediaResponse(media), nil
}

func (server *Server) GetPost(ctx context.Context, request *contentv1.GetPostRequest) (*contentv1.PostResponse, error) {
	post, err := server.contentUseCase.GetPost(ctx, mappers.GetPostRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostResponse(post), nil
}

func (server *Server) UpdatePost(ctx context.Context, request *contentv1.UpdatePostRequest) (*contentv1.PostResponse, error) {
	post, err := server.contentUseCase.UpdatePost(ctx, mappers.UpdatePostRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostResponse(post), nil
}

func (server *Server) DeletePost(ctx context.Context, request *contentv1.DeletePostRequest) (*emptypb.Empty, error) {
	if err := server.contentUseCase.DeletePost(ctx, mappers.DeletePostRequestToCommand(request)); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.Empty(), nil
}

func (server *Server) LikePost(ctx context.Context, request *contentv1.LikePostRequest) (*contentv1.PostLikeStateResponse, error) {
	state, err := server.contentUseCase.LikePost(ctx, mappers.LikePostRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostLikeStateResponse(state), nil
}

func (server *Server) UnlikePost(ctx context.Context, request *contentv1.UnlikePostRequest) (*contentv1.PostLikeStateResponse, error) {
	state, err := server.contentUseCase.UnlikePost(ctx, mappers.UnlikePostRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewPostLikeStateResponse(state), nil
}

func (server *Server) CreateComment(ctx context.Context, request *contentv1.CreateCommentRequest) (*contentv1.CommentResponse, error) {
	comment, err := server.contentUseCase.CreateComment(ctx, mappers.CreateCommentRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewCommentResponse(comment), nil
}

func (server *Server) ListComments(ctx context.Context, request *contentv1.ListCommentsRequest) (*contentv1.ListCommentsResponse, error) {
	comments, err := server.contentUseCase.ListComments(ctx, mappers.ListCommentsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListCommentsResponse(comments), nil
}
