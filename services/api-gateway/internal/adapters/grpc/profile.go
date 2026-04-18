package grpc

import (
	"context"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/mappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) CreateProfile(ctx context.Context, request *gatewayv1.CreateProfileRequest) (*gatewayv1.ProfileResponse, error) {
	response, err := server.profileClient.CreateProfile(forwardContext(ctx), mappers.CreateProfileRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.ProfileResponseFromProfile(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map create profile response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) GetProfile(ctx context.Context, request *gatewayv1.GetProfileRequest) (*gatewayv1.ProfileResponse, error) {
	response, err := server.profileClient.GetProfile(forwardContext(ctx), mappers.GetProfileRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.ProfileResponseFromProfile(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map get profile response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) UpdateProfile(ctx context.Context, request *gatewayv1.UpdateProfileRequest) (*gatewayv1.ProfileResponse, error) {
	response, err := server.profileClient.UpdateProfile(forwardContext(ctx), mappers.UpdateProfileRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.ProfileResponseFromProfile(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map update profile response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) SearchAuthors(ctx context.Context, request *gatewayv1.SearchAuthorsRequest) (*gatewayv1.SearchAuthorsResponse, error) {
	response, err := server.profileClient.SearchAuthors(forwardContext(ctx), mappers.SearchAuthorsRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.SearchAuthorsResponseFromProfile(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map search authors response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) UploadAvatar(ctx context.Context, request *gatewayv1.UploadAvatarRequest) (*gatewayv1.ProfileResponse, error) {
	response, err := server.profileClient.UploadAvatar(forwardContext(ctx), mappers.UploadAvatarRequestToProfile(request))
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.ProfileResponseFromProfile(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map upload avatar response: %v", err)
	}

	return mappedResponse, nil
}

func (server *Server) DeleteAvatar(ctx context.Context, request *gatewayv1.DeleteAvatarRequest) (*emptypb.Empty, error) {
	return server.profileClient.DeleteAvatar(forwardContext(ctx), mappers.DeleteAvatarRequestToProfile(request))
}

func (server *Server) ListSportTypes(ctx context.Context, _ *emptypb.Empty) (*gatewayv1.ListSportTypesResponse, error) {
	response, err := server.profileClient.ListSportTypes(forwardContext(ctx), &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	mappedResponse, err := mappers.ListSportTypesResponseFromProfile(response)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "map list sport types response: %v", err)
	}

	return mappedResponse, nil
}
