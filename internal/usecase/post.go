package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

var ErrPostNotFound = errors.New("post not found")
var ErrPostForbidden = errors.New("post forbidden")

type PostUseCase struct {
	postRepository postRepository
}

func NewPostUseCase(postRepository postRepository) *PostUseCase {
	return &PostUseCase{
		postRepository: postRepository,
	}
}

func (useCase *PostUseCase) ListProfilePosts(ctx context.Context, profileUserID int64, currentUserID int64) ([]domain.PostListItem, error) {
	return useCase.postRepository.ListProfilePosts(ctx, profileUserID, currentUserID)
}

func (useCase *PostUseCase) GetByID(ctx context.Context, postID int64, currentUserID int64) (domain.Post, error) {
	post, err := useCase.postRepository.GetByID(ctx, postID, currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Post{}, ErrPostNotFound
		}

		return domain.Post{}, err
	}

	if !post.CanView {
		return domain.Post{}, ErrPostForbidden
	}

	return post, nil
}

func (useCase *PostUseCase) SetLike(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error) {
	postLikeStatus, err := useCase.postRepository.SetLike(ctx, postID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PostLikeStatus{}, ErrPostNotFound
		}

		return domain.PostLikeStatus{}, err
	}

	return postLikeStatus, nil
}

func (useCase *PostUseCase) DeleteLike(ctx context.Context, postID int64, userID int64) (domain.PostLikeStatus, error) {
	postLikeStatus, err := useCase.postRepository.DeleteLike(ctx, postID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PostLikeStatus{}, ErrPostNotFound
		}

		return domain.PostLikeStatus{}, err
	}

	return postLikeStatus, nil
}
