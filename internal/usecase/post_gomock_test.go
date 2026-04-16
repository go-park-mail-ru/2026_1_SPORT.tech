package usecase_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/gen"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/usecase"
	"github.com/golang/mock/gomock"
)

func TestPostUseCaseGetByID(t *testing.T) {
	t.Run("maps not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		repository.EXPECT().GetByID(gomock.Any(), int64(10), int64(2)).Return(domain.Post{}, sql.ErrNoRows)

		useCase := usecase.NewPostUseCase(repository)

		_, err := useCase.GetByID(context.Background(), 10, 2)
		if !errors.Is(err, usecase.ErrPostNotFound) {
			t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrPostNotFound)
		}
	})

	t.Run("maps forbidden", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		repository.EXPECT().GetByID(gomock.Any(), int64(10), int64(2)).Return(domain.Post{CanView: false}, nil)

		useCase := usecase.NewPostUseCase(repository)

		_, err := useCase.GetByID(context.Background(), 10, 2)
		if !errors.Is(err, usecase.ErrPostForbidden) {
			t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrPostForbidden)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		expected := domain.Post{PostID: 10, TrainerID: 2, CanView: true}
		repository.EXPECT().GetByID(gomock.Any(), int64(10), int64(2)).Return(expected, nil)

		useCase := usecase.NewPostUseCase(repository)

		post, err := useCase.GetByID(context.Background(), 10, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if post.PostID != expected.PostID || post.TrainerID != expected.TrainerID {
			t.Fatalf("unexpected post: got %+v, expect %+v", post, expected)
		}
	})
}

func TestPostUseCaseCreateUpdateDeleteAndLikes(t *testing.T) {
	t.Run("create success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		command := usecase.CreatePostCommand{Title: "hello", TextContent: "world"}
		expected := domain.Post{PostID: 20, TrainerID: 7, CanView: true}

		repository.EXPECT().Create(gomock.Any(), int64(7), command).Return(int64(20), nil)
		repository.EXPECT().GetByID(gomock.Any(), int64(20), int64(7)).Return(expected, nil)

		useCase := usecase.NewPostUseCase(repository)

		post, err := useCase.Create(context.Background(), 7, command)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if post.PostID != expected.PostID || post.Title != expected.Title {
			t.Fatalf("unexpected post: got %+v, expect %+v", post, expected)
		}
	})

	t.Run("create maps tier error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		repository.EXPECT().Create(gomock.Any(), int64(7), gomock.Any()).Return(int64(0), usecase.ErrPostTierNotFound)

		useCase := usecase.NewPostUseCase(repository)

		_, err := useCase.Create(context.Background(), 7, usecase.CreatePostCommand{})
		if !errors.Is(err, usecase.ErrPostTierNotFound) {
			t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrPostTierNotFound)
		}
	})

	t.Run("update success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		title := "updated"
		command := usecase.UpdatePostCommand{Title: &title}
		expected := domain.Post{PostID: 21, TrainerID: 7, CanView: true, Title: title}

		repository.EXPECT().Update(gomock.Any(), int64(7), int64(21), command).Return(nil)
		repository.EXPECT().GetByID(gomock.Any(), int64(21), int64(7)).Return(expected, nil)

		useCase := usecase.NewPostUseCase(repository)

		post, err := useCase.Update(context.Background(), 7, 21, command)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if post.PostID != expected.PostID || post.Title != expected.Title {
			t.Fatalf("unexpected post: got %+v, expect %+v", post, expected)
		}
	})

	t.Run("update maps errors", func(t *testing.T) {
		tests := []struct {
			name      string
			repoErr   error
			expectErr error
		}{
			{name: "not found", repoErr: sql.ErrNoRows, expectErr: usecase.ErrPostNotFound},
			{name: "forbidden", repoErr: usecase.ErrPostForbidden, expectErr: usecase.ErrPostForbidden},
			{name: "tier", repoErr: usecase.ErrPostTierNotFound, expectErr: usecase.ErrPostTierNotFound},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				repository := gen.NewMockpostRepository(ctrl)
				repository.EXPECT().Update(gomock.Any(), int64(7), int64(21), gomock.Any()).Return(tt.repoErr)

				useCase := usecase.NewPostUseCase(repository)

				_, err := useCase.Update(context.Background(), 7, 21, usecase.UpdatePostCommand{})
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
				}
			})
		}
	})

	t.Run("delete maps errors", func(t *testing.T) {
		tests := []struct {
			name      string
			repoErr   error
			expectErr error
		}{
			{name: "not found", repoErr: sql.ErrNoRows, expectErr: usecase.ErrPostNotFound},
			{name: "forbidden", repoErr: usecase.ErrPostForbidden, expectErr: usecase.ErrPostForbidden},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
				repository := gen.NewMockpostRepository(ctrl)
				repository.EXPECT().Delete(gomock.Any(), int64(7), int64(21)).Return(tt.repoErr)

				useCase := usecase.NewPostUseCase(repository)

				err := useCase.Delete(context.Background(), 7, 21)
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("unexpected error: got %v, expect %v", err, tt.expectErr)
				}
			})
		}
	})

	t.Run("set like success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		expected := domain.PostLikeStatus{PostID: 21, LikesCount: 3, IsLiked: true}
		repository.EXPECT().SetLike(gomock.Any(), int64(21), int64(7)).Return(expected, nil)

		useCase := usecase.NewPostUseCase(repository)

		status, err := useCase.SetLike(context.Background(), 21, 7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status != expected {
			t.Fatalf("unexpected status: got %+v, expect %+v", status, expected)
		}
	})

	t.Run("delete like not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		repository.EXPECT().DeleteLike(gomock.Any(), int64(21), int64(7)).Return(domain.PostLikeStatus{}, sql.ErrNoRows)

		useCase := usecase.NewPostUseCase(repository)

		_, err := useCase.DeleteLike(context.Background(), 21, 7)
		if !errors.Is(err, usecase.ErrPostNotFound) {
			t.Fatalf("unexpected error: got %v, expect %v", err, usecase.ErrPostNotFound)
		}
	})

	t.Run("list profile posts passthrough", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		repository := gen.NewMockpostRepository(ctrl)
		expected := []domain.PostListItem{{PostID: 1, Title: "one", CreatedAt: time.Now()}}
		repository.EXPECT().ListProfilePosts(gomock.Any(), int64(3), int64(7)).Return(expected, nil)

		useCase := usecase.NewPostUseCase(repository)

		posts, err := useCase.ListProfilePosts(context.Background(), 3, 7)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(posts) != 1 || posts[0].PostID != expected[0].PostID {
			t.Fatalf("unexpected posts: got %+v, expect %+v", posts, expected)
		}
	})
}
