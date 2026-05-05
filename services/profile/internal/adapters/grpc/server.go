package grpc

import (
	"context"

	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/mappers"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProfileUseCase interface {
	CreateProfile(ctx context.Context, command usecase.CreateProfileCommand) (domain.Profile, error)
	GetProfile(ctx context.Context, userID int64) (domain.Profile, error)
	UpdateProfile(ctx context.Context, command usecase.UpdateProfileCommand) (domain.Profile, error)
}

type AuthorUseCase interface {
	SearchAuthors(ctx context.Context, query usecase.SearchAuthorsQuery) ([]domain.AuthorSummary, error)
}

type AvatarUseCase interface {
	UploadAvatar(ctx context.Context, command usecase.UploadAvatarCommand) (domain.Profile, error)
	DeleteAvatar(ctx context.Context, userID int64) error
}

type SportUseCase interface {
	ListSportTypes(ctx context.Context) ([]domain.SportType, error)
}

type UseCases struct {
	Profiles ProfileUseCase
	Authors  AuthorUseCase
	Avatars  AvatarUseCase
	Sports   SportUseCase
}

type Server struct {
	profilev1.UnimplementedProfileServiceServer
	useCases UseCases
}

func NewServer(useCases UseCases) *Server {
	return &Server{useCases: useCases}
}

func (server *Server) CreateProfile(ctx context.Context, request *profilev1.CreateProfileRequest) (*profilev1.ProfileResponse, error) {
	profile, err := server.useCases.Profiles.CreateProfile(ctx, mappers.CreateProfileRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewProfileResponse(profile), nil
}

func (server *Server) GetProfile(ctx context.Context, request *profilev1.GetProfileRequest) (*profilev1.ProfileResponse, error) {
	profile, err := server.useCases.Profiles.GetProfile(ctx, request.GetUserId())
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewProfileResponse(profile), nil
}

func (server *Server) UpdateProfile(ctx context.Context, request *profilev1.UpdateProfileRequest) (*profilev1.ProfileResponse, error) {
	profile, err := server.useCases.Profiles.UpdateProfile(ctx, mappers.UpdateProfileRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewProfileResponse(profile), nil
}

func (server *Server) SearchAuthors(ctx context.Context, request *profilev1.SearchAuthorsRequest) (*profilev1.SearchAuthorsResponse, error) {
	authors, err := server.useCases.Authors.SearchAuthors(ctx, mappers.SearchAuthorsRequestToQuery(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewSearchAuthorsResponse(authors), nil
}

func (server *Server) UploadAvatar(ctx context.Context, request *profilev1.UploadAvatarRequest) (*profilev1.ProfileResponse, error) {
	profile, err := server.useCases.Avatars.UploadAvatar(ctx, mappers.UploadAvatarRequestToCommand(request))
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewProfileResponse(profile), nil
}

func (server *Server) DeleteAvatar(ctx context.Context, request *profilev1.DeleteAvatarRequest) (*emptypb.Empty, error) {
	if err := server.useCases.Avatars.DeleteAvatar(ctx, request.GetUserId()); err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.Empty(), nil
}

func (server *Server) ListSportTypes(ctx context.Context, request *emptypb.Empty) (*profilev1.ListSportTypesResponse, error) {
	sportTypes, err := server.useCases.Sports.ListSportTypes(ctx)
	if err != nil {
		return nil, mappers.ErrorToStatus(err)
	}

	return mappers.NewListSportTypesResponse(sportTypes), nil
}
