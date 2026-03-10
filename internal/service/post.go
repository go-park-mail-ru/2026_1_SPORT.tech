package service

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/repository"
)

var ErrPostNotFound = errors.New("post not found")
var ErrPostForbidden = errors.New("post forbidden")

type postRepository interface {
	ListProfilePosts(ctx context.Context, profileUserID int64, currentUserID int64) ([]repository.PostListItem, error)
	GetByID(ctx context.Context, postID int64, currentUserID int64) (repository.Post, error)
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

type PostAttachment struct {
	PostAttachmentID int64
	Kind             string
	FileURL          string
}

type Post struct {
	PostID      int64
	TrainerID   int64
	MinTierID   *int64
	Title       string
	TextContent string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Attachments []PostAttachment
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

func (service *PostService) GetByID(ctx context.Context, postID int64, currentUserID int64) (Post, error) {
	post, err := service.postRepository.GetByID(ctx, postID, currentUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Post{}, ErrPostNotFound
		}

		return Post{}, err
	}

	if !post.CanView {
		return Post{}, ErrPostForbidden
	}

	attachments := make([]PostAttachment, 0, len(post.Attachments))
	for _, attachment := range post.Attachments {
		attachments = append(attachments, PostAttachment{
			PostAttachmentID: attachment.PostAttachmentID,
			Kind:             attachment.Kind,
			FileURL:          attachment.FileURL,
		})
	}

	return Post{
		PostID:      post.PostID,
		TrainerID:   post.TrainerID,
		MinTierID:   post.MinTierID,
		Title:       post.Title,
		TextContent: post.TextContent,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Attachments: attachments,
	}, nil
}
