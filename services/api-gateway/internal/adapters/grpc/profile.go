package grpc

import (
	"context"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) GetProfile(ctx context.Context, request *gatewayv1.GetProfileRequest) (*gatewayv1.ProfileResponse, error) {
	principal, err := server.optionalSession(ctx)
	if err != nil {
		return nil, err
	}

	response, err := server.profileClient.GetProfile(
		forwardContext(ctx),
		&profilev1.GetProfileRequest{UserId: int64(request.GetUserId())},
	)
	if err != nil {
		return nil, err
	}

	var currentUserID int64
	if principal != nil && principal.User != nil {
		currentUserID = principal.User.GetUserId()
	}

	return mappers.ProfileResponseFromProfile(response.GetProfile(), currentUserID)
}

func (server *Server) ListTrainers(ctx context.Context, request *gatewayv1.ListTrainersRequest) (*gatewayv1.GetTrainersResponse, error) {
	return server.searchTrainers(ctx, request)
}

func (server *Server) SearchTrainers(ctx context.Context, request *gatewayv1.ListTrainersRequest) (*gatewayv1.GetTrainersResponse, error) {
	return server.searchTrainers(ctx, request)
}

func (server *Server) searchTrainers(ctx context.Context, request *gatewayv1.ListTrainersRequest) (*gatewayv1.GetTrainersResponse, error) {
	response, err := server.profileClient.SearchAuthors(
		forwardContext(ctx),
		mappers.ListTrainersRequestToProfile(request),
	)
	if err != nil {
		return nil, err
	}

	return mappers.GetTrainersResponseFromProfile(response)
}

func (server *Server) UpdateMyProfile(ctx context.Context, request *gatewayv1.UpdateMyProfileRequest) (*gatewayv1.ProfileResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	updateRequest, err := mappers.UpdateMyProfileRequestToProfile(userID, request)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	response, err := server.profileClient.UpdateProfile(forwardContext(ctx), updateRequest)
	if err != nil {
		return nil, err
	}

	return mappers.ProfileResponseFromProfile(response.GetProfile(), userID)
}

func (server *Server) UploadMyAvatar(ctx context.Context, request *gatewayv1.UploadMyAvatarRequest) (*gatewayv1.AvatarUploadResponse, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.profileClient.UploadAvatar(
		forwardContext(ctx),
		mappers.UploadMyAvatarRequestToProfile(userID, request),
	)
	if err != nil {
		return nil, err
	}

	return mappers.AvatarUploadResponseFromProfile(response.GetProfile()), nil
}

func (server *Server) DeleteMyAvatar(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	principal, err := server.requireSession(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := userIDFromPrincipal(principal)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if _, err := server.profileClient.DeleteAvatar(
		forwardContext(ctx),
		&profilev1.DeleteAvatarRequest{UserId: userID},
	); err != nil {
		return nil, err
	}

	if err := setHTTPStatus(ctx, 204); err != nil {
		return nil, status.Errorf(codes.Internal, "set response status: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (server *Server) ListProfilePosts(ctx context.Context, request *gatewayv1.GetProfileRequest) (*gatewayv1.ProfilePostsResponse, error) {
	principal, err := server.optionalSession(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := server.profileClient.GetProfile(
		forwardContext(ctx),
		&profilev1.GetProfileRequest{UserId: int64(request.GetUserId())},
	); err != nil {
		return nil, err
	}

	var viewerUserID int64
	if principal != nil && principal.User != nil {
		viewerUserID = principal.User.GetUserId()
	}

	response, err := server.contentClient.ListAuthorPosts(
		forwardContext(ctx),
		&contentv1.ListAuthorPostsRequest{
			AuthorUserId: int64(request.GetUserId()),
			ViewerUserId: viewerUserID,
			Limit:        20,
			Offset:       0,
		},
	)
	if err != nil {
		return nil, err
	}

	return mappers.ProfilePostsResponseFromContent(request.GetUserId(), response)
}
