package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/domain"
)

var ErrPostNotFound = errors.New("post not found")
var ErrPostForbidden = errors.New("post forbidden")
var ErrPostTierNotFound = errors.New("post tier not found")

type CreatePostAttachmentCommand struct {
	Kind    string
	FileURL string
}

type CreatePostCommand struct {
	MinTierID   *int64
	Title       string
	TextContent string
	Attachments []CreatePostAttachmentCommand
}

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

func (useCase *PostUseCase) Create(ctx context.Context, trainerID int64, command CreatePostCommand) (domain.Post, error) {
	postID, err := useCase.postRepository.Create(ctx, trainerID, command)
	if err != nil {
		if errors.Is(err, ErrPostTierNotFound) {
			return domain.Post{}, ErrPostTierNotFound
		}

		return domain.Post{}, err
	}

	return useCase.GetByID(ctx, postID, trainerID)
}
