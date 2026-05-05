package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PostUseCase interface {
	ListAuthorPosts(ctx context.Context, query usecase.ListAuthorPostsQuery) ([]domain.PostSummary, error)
	SearchPosts(ctx context.Context, query usecase.SearchPostsQuery) ([]domain.PostSummary, error)
	CreatePost(ctx context.Context, command usecase.CreatePostCommand) (domain.Post, error)
	GetPost(ctx context.Context, query usecase.GetPostQuery) (domain.Post, error)
	UpdatePost(ctx context.Context, command usecase.UpdatePostCommand) (domain.Post, error)
	DeletePost(ctx context.Context, command usecase.DeletePostCommand) error
	LikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
	UnlikePost(ctx context.Context, command usecase.LikePostCommand) (domain.PostLikeState, error)
}

type PostMediaUseCase interface {
	UploadPostMedia(ctx context.Context, command usecase.UploadPostMediaCommand) (domain.PostMedia, error)
}

type TierUseCase interface {
	ListSubscriptionTiers(ctx context.Context, query usecase.ListSubscriptionTiersQuery) ([]domain.SubscriptionTier, error)
	CreateSubscriptionTier(ctx context.Context, command usecase.CreateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	UpdateSubscriptionTier(ctx context.Context, command usecase.UpdateSubscriptionTierCommand) (domain.SubscriptionTier, error)
	DeleteSubscriptionTier(ctx context.Context, command usecase.DeleteSubscriptionTierCommand) error
}

type SubscriptionUseCase interface {
	SubscribeToTrainer(ctx context.Context, command usecase.SubscribeToTrainerCommand) (domain.Subscription, error)
	ListMySubscriptions(ctx context.Context, query usecase.ListMySubscriptionsQuery) ([]domain.Subscription, error)
	UpdateSubscription(ctx context.Context, command usecase.UpdateSubscriptionCommand) (domain.Subscription, error)
	CancelSubscription(ctx context.Context, command usecase.CancelSubscriptionCommand) error
}

type CommentUseCase interface {
	CreateComment(ctx context.Context, command usecase.CreateCommentCommand) (domain.Comment, error)
	ListComments(ctx context.Context, query usecase.ListCommentsQuery) ([]domain.Comment, error)
}

type UseCases struct {
	Posts         PostUseCase
	PostMedia     PostMediaUseCase
	Tiers         TierUseCase
	Subscriptions SubscriptionUseCase
	Comments      CommentUseCase
}

type Server struct {
	contentv1.UnimplementedContentServiceServer
	useCases UseCases
}

func NewServer(useCases UseCases) *Server {
	return &Server{useCases: useCases}
}

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

func (server *Server) ListSubscriptionTiers(ctx context.Context, request *contentv1.ListSubscriptionTiersRequest) (*contentv1.ListSubscriptionTiersResponse, error) {
	tiers, err := server.useCases.Tiers.ListSubscriptionTiers(ctx, mappers.ListSubscriptionTiersRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListSubscriptionTiersResponse(tiers), nil
}

func (server *Server) CreateSubscriptionTier(ctx context.Context, request *contentv1.CreateSubscriptionTierRequest) (*contentv1.SubscriptionTier, error) {
	tier, err := server.useCases.Tiers.CreateSubscriptionTier(ctx, mappers.CreateSubscriptionTierRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSubscriptionTierResponse(tier), nil
}

func (server *Server) UpdateSubscriptionTier(ctx context.Context, request *contentv1.UpdateSubscriptionTierRequest) (*contentv1.SubscriptionTier, error) {
	tier, err := server.useCases.Tiers.UpdateSubscriptionTier(ctx, mappers.UpdateSubscriptionTierRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSubscriptionTierResponse(tier), nil
}

func (server *Server) DeleteSubscriptionTier(ctx context.Context, request *contentv1.DeleteSubscriptionTierRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Tiers.DeleteSubscriptionTier(ctx, mappers.DeleteSubscriptionTierRequestToCommand(request)); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.Empty(), nil
}

func (server *Server) SubscribeToTrainer(ctx context.Context, request *contentv1.SubscribeToTrainerRequest) (*contentv1.Subscription, error) {
	subscription, err := server.useCases.Subscriptions.SubscribeToTrainer(ctx, mappers.SubscribeToTrainerRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSubscriptionResponse(subscription), nil
}

func (server *Server) ListMySubscriptions(ctx context.Context, request *contentv1.ListMySubscriptionsRequest) (*contentv1.ListMySubscriptionsResponse, error) {
	subscriptions, err := server.useCases.Subscriptions.ListMySubscriptions(ctx, mappers.ListMySubscriptionsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListMySubscriptionsResponse(subscriptions), nil
}

func (server *Server) UpdateSubscription(ctx context.Context, request *contentv1.UpdateSubscriptionRequest) (*contentv1.Subscription, error) {
	subscription, err := server.useCases.Subscriptions.UpdateSubscription(ctx, mappers.UpdateSubscriptionRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSubscriptionResponse(subscription), nil
}

func (server *Server) CancelSubscription(ctx context.Context, request *contentv1.CancelSubscriptionRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Subscriptions.CancelSubscription(ctx, mappers.CancelSubscriptionRequestToCommand(request)); err != nil {
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
