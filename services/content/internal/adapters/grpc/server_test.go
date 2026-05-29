package grpc_test

import (
	"context"
	"testing"
	"time"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/mocks"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerGetPost(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	ctrl := gomock.NewController(t)
	contentUseCase := mocks.NewMockContentUseCase(ctrl)
	contentUseCase.EXPECT().
		GetPost(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, query usecase.GetPostQuery) (domain.Post, error) {
			return domain.Post{
				PostID:       query.PostID,
				AuthorUserID: 7,
				Title:        "Morning run",
				CreatedAt:    now,
				UpdatedAt:    now,
				CanView:      true,
			}, nil
		})

	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Posts:         contentUseCase,
		PostMedia:     contentUseCase,
		Tiers:         contentUseCase,
		Subscriptions: contentUseCase,
		Comments:      contentUseCase,
		Donations:     contentUseCase,
	})

	response, err := server.GetPost(context.Background(), &contentv1.GetPostRequest{PostId: 7, ViewerUserId: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetPost().GetPostId() != 7 {
		t.Fatalf("unexpected post id: %d", response.GetPost().GetPostId())
	}
}

func TestServerGetPostMapsForbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	contentUseCase := mocks.NewMockContentUseCase(ctrl)
	contentUseCase.EXPECT().
		GetPost(gomock.Any(), gomock.Any()).
		Return(domain.Post{}, domain.ErrPostForbidden)

	server := grpcadapter.NewServer(grpcadapter.UseCases{
		Posts:         contentUseCase,
		PostMedia:     contentUseCase,
		Tiers:         contentUseCase,
		Subscriptions: contentUseCase,
		Comments:      contentUseCase,
		Donations:     contentUseCase,
	})

	_, err := server.GetPost(context.Background(), &contentv1.GetPostRequest{PostId: 7, ViewerUserId: 3})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("unexpected status code: %s", status.Code(err))
	}
}
