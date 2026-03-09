package service

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
)

type postRepository interface {
	ListProfilePosts(ctx context.Context, profileUserID int64, currentUserID int64) ([]repository.PostListItem, error)
}

type PostListItem struct {
	PostID    int64
	TrainerID int64
	MinTierID *int64
	Title     string
	CreatedAt time.Time
	CanView   bool
}

type PostService struct {
	postRepository postRepository
}

func NewPostService(postRepository postRepository) *PostService {
	return &PostService{
		postRepository: postRepository,
	}
}

func (service *PostService) ListProfilePosts(ctx context.Context, profileUserID int64, currentUserID int64) ([]PostListItem, error) {
	posts, err := service.postRepository.ListProfilePosts(ctx, profileUserID, currentUserID)
	if err != nil {
		return nil, err
	}

	result := make([]PostListItem, 0, len(posts))
	for _, post := range posts {
		result = append(result, PostListItem{
			PostID:    post.PostID,
			TrainerID: post.TrainerID,
			MinTierID: post.MinTierID,
			Title:     post.Title,
			CreatedAt: post.CreatedAt,
			CanView:   post.CanView,
		})
	}

	return result, nil
}
