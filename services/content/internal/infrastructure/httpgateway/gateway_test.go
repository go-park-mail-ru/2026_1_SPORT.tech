package httpgateway_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/httpgateway"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/mocks"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"go.uber.org/mock/gomock"
)

func TestNewLocalMuxExposesGeneratedGetPostEndpoint(t *testing.T) {
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

	handler := grpcadapter.NewServer(grpcadapter.UseCases{
		Posts:         contentUseCase,
		PostMedia:     contentUseCase,
		Tiers:         contentUseCase,
		Subscriptions: contentUseCase,
		Comments:      contentUseCase,
		Donations:     contentUseCase,
	})

	mux, err := httpgateway.NewLocalMux(context.Background(), handler)
	if err != nil {
		t.Fatalf("new local mux: %v", err)
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	response, err := http.Get(server.URL + "/v1/posts/7?viewerUserId=7")
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d", response.StatusCode)
	}

	var payload struct {
		Post struct {
			PostID string `json:"postId"`
		} `json:"post"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Post.PostID != "7" {
		t.Fatalf("unexpected post id: %s", payload.Post.PostID)
	}
}
