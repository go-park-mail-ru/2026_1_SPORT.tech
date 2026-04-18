package grpc

import (
	"context"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	gatewayv1.UnimplementedAuthServiceServer
	gatewayv1.UnimplementedProfileServiceServer
	gatewayv1.UnimplementedContentServiceServer
	authClient    authv1.AuthServiceClient
	profileClient profilev1.ProfileServiceClient
	contentClient contentv1.ContentServiceClient
}

func NewServer(
	authClient authv1.AuthServiceClient,
	profileClient profilev1.ProfileServiceClient,
	contentClient contentv1.ContentServiceClient,
) *Server {
	return &Server{
		authClient:    authClient,
		profileClient: profileClient,
		contentClient: contentClient,
	}
}

func (server *Server) Register(ctx context.Context, request *gatewayv1.RegisterRequest) (*gatewayv1.AuthSessionResponse, error) {
	response, err := server.authClient.Register(forwardContext(ctx), mappers.RegisterRequestToAuth(request))
	if err != nil {
		return nil, err
	}

	return mappers.AuthSessionResponseFromAuth(response), nil
}

func (server *Server) Login(ctx context.Context, request *gatewayv1.LoginRequest) (*gatewayv1.AuthSessionResponse, error) {
	response, err := server.authClient.Login(forwardContext(ctx), mappers.LoginRequestToAuth(request))
	if err != nil {
		return nil, err
	}

	return mappers.AuthSessionResponseFromAuth(response), nil
}

func (server *Server) Logout(ctx context.Context, request *gatewayv1.LogoutRequest) (*emptypb.Empty, error) {
	return server.authClient.Logout(forwardContext(ctx), mappers.LogoutRequestToAuth(request))
}

func (server *Server) ResolveSession(ctx context.Context, request *gatewayv1.ResolveSessionRequest) (*gatewayv1.ResolveSessionResponse, error) {
	response, err := server.authClient.GetSession(forwardContext(ctx), mappers.ResolveSessionRequestToAuth(request))
	if err != nil {
		return nil, err
	}

	return mappers.ResolveSessionResponseFromAuth(response), nil
}

func (server *Server) CreateProfile(ctx context.Context, request *gatewayv1.CreateProfileRequest) (*gatewayv1.ProfileResponse, error) {
	response, err := server.profileClient.CreateProfile(forwardContext(ctx), mappers.CreateProfileRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	return mappers.ProfileResponseFromProfile(response), nil
}

func (server *Server) GetProfile(ctx context.Context, request *gatewayv1.GetProfileRequest) (*gatewayv1.ProfileResponse, error) {
	response, err := server.profileClient.GetProfile(forwardContext(ctx), mappers.GetProfileRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	return mappers.ProfileResponseFromProfile(response), nil
}

func (server *Server) UpdateProfile(ctx context.Context, request *gatewayv1.UpdateProfileRequest) (*gatewayv1.ProfileResponse, error) {
	response, err := server.profileClient.UpdateProfile(forwardContext(ctx), mappers.UpdateProfileRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	return mappers.ProfileResponseFromProfile(response), nil
}

func (server *Server) SearchAuthors(ctx context.Context, request *gatewayv1.SearchAuthorsRequest) (*gatewayv1.SearchAuthorsResponse, error) {
	response, err := server.profileClient.SearchAuthors(forwardContext(ctx), mappers.SearchAuthorsRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	return mappers.SearchAuthorsResponseFromProfile(response), nil
}

func (server *Server) UploadAvatar(ctx context.Context, request *gatewayv1.UploadAvatarRequest) (*gatewayv1.ProfileResponse, error) {
	response, err := server.profileClient.UploadAvatar(forwardContext(ctx), mappers.UploadAvatarRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	return mappers.ProfileResponseFromProfile(response), nil
}

func (server *Server) DeleteAvatar(ctx context.Context, request *gatewayv1.DeleteAvatarRequest) (*emptypb.Empty, error) {
	return server.profileClient.DeleteAvatar(forwardContext(ctx), mappers.DeleteAvatarRequestToProfile(request))
}

func (server *Server) ListSportTypes(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.ListSportTypesResponse, error) {
	response, err := server.profileClient.ListSportTypes(forwardContext(ctx), &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	return mappers.ListSportTypesResponseFromProfile(response), nil
}

func (server *Server) ListAuthorPosts(ctx context.Context, request *gatewayv1.ListAuthorPostsRequest) (*gatewayv1.ListAuthorPostsResponse, error) {
	response, err := server.contentClient.ListAuthorPosts(forwardContext(ctx), mappers.ListAuthorPostsRequestToContent(request))
	if err != nil {
		return nil, err
	}

	return mappers.ListAuthorPostsResponseFromContent(response), nil
}

func (server *Server) CreatePost(ctx context.Context, request *gatewayv1.CreatePostRequest) (*gatewayv1.PostResponse, error) {
	response, err := server.contentClient.CreatePost(forwardContext(ctx), mappers.CreatePostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	return mappers.PostResponseFromContent(response), nil
}

func (server *Server) GetPost(ctx context.Context, request *gatewayv1.GetPostRequest) (*gatewayv1.PostResponse, error) {
	response, err := server.contentClient.GetPost(forwardContext(ctx), mappers.GetPostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	return mappers.PostResponseFromContent(response), nil
}

func (server *Server) UpdatePost(ctx context.Context, request *gatewayv1.UpdatePostRequest) (*gatewayv1.PostResponse, error) {
	response, err := server.contentClient.UpdatePost(forwardContext(ctx), mappers.UpdatePostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	return mappers.PostResponseFromContent(response), nil
}

func (server *Server) DeletePost(ctx context.Context, request *gatewayv1.DeletePostRequest) (*emptypb.Empty, error) {
	return server.contentClient.DeletePost(forwardContext(ctx), mappers.DeletePostRequestToContent(request))
}

func (server *Server) LikePost(ctx context.Context, request *gatewayv1.LikePostRequest) (*gatewayv1.PostLikeStateResponse, error) {
	response, err := server.contentClient.LikePost(forwardContext(ctx), mappers.LikePostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	return mappers.PostLikeStateResponseFromContent(response), nil
}

func (server *Server) UnlikePost(ctx context.Context, request *gatewayv1.UnlikePostRequest) (*gatewayv1.PostLikeStateResponse, error) {
	response, err := server.contentClient.UnlikePost(forwardContext(ctx), mappers.UnlikePostRequestToContent(request))
	if err != nil {
		return nil, err
	}

	return mappers.PostLikeStateResponseFromContent(response), nil
}

func (server *Server) CreateComment(ctx context.Context, request *gatewayv1.CreateCommentRequest) (*gatewayv1.CommentResponse, error) {
	response, err := server.contentClient.CreateComment(forwardContext(ctx), mappers.CreateCommentRequestToContent(request))
	if err != nil {
		return nil, err
	}

	return mappers.CommentResponseFromContent(response), nil
}

func (server *Server) ListComments(ctx context.Context, request *gatewayv1.ListCommentsRequest) (*gatewayv1.ListCommentsResponse, error) {
	response, err := server.contentClient.ListComments(forwardContext(ctx), mappers.ListCommentsRequestToContent(request))
	if err != nil {
		return nil, err
	}

	return mappers.ListCommentsResponseFromContent(response), nil
}

func forwardContext(ctx context.Context) context.Context {
	incomingMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	outgoingMD := metadata.MD{}
	for _, key := range []string{"authorization", "x-request-id", "x-session-token", "x-user-id", "x-subscription-level"} {
		values := incomingMD.Get(key)
		if len(values) == 0 {
			continue
		}

		copied := append([]string(nil), values...)
		outgoingMD.Set(key, copied...)
	}

	if len(outgoingMD) == 0 {
		return ctx
	}

	return metadata.NewOutgoingContext(ctx, outgoingMD)
}
