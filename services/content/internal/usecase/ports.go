package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

type ContentRepository interface {
	CreatePost(ctx context.Context, post domain.Post) (int64, error)
	GetPost(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error)
	ListAuthorPosts(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error)
	SearchPosts(ctx context.Context, query SearchPostsQuery) ([]domain.PostSummary, error)
	UpdatePost(ctx context.Context, post domain.Post, replaceBlocks bool) error
	DeletePost(ctx context.Context, postID int64, authorUserID int64) error
	ListSubscriptionTiers(ctx context.Context, trainerUserID int64) ([]domain.SubscriptionTier, error)
	GetSubscriptionTier(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error)
	CreateSubscriptionTier(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error)
	UpdateSubscriptionTier(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error)
	DeleteSubscriptionTier(ctx context.Context, trainerUserID int64, tierID int64) error
	GetActiveSubscriptionLevel(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error)
	SubscribeToTrainer(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	ListSubscriptions(ctx context.Context, clientUserID int64) ([]domain.Subscription, error)
	CancelSubscription(ctx context.Context, clientUserID int64, subscriptionID int64) error
	UpsertLike(ctx context.Context, postID int64, userID int64) error
	DeleteLike(ctx context.Context, postID int64, userID int64) error
	GetPostLikeState(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error)
	CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	ListComments(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error)
}

type PostMediaStorage interface {
	UploadPostMedia(ctx context.Context, authorUserID int64, fileName string, contentType string, file io.Reader, size int64) (string, error)
}

type PostBlockInput struct {
	Kind        domain.BlockKind
	TextContent *string
	FileURL     *string
}

type ListAuthorPostsQuery struct {
	AuthorUserID            int64
	ViewerUserID            int64
	ViewerSubscriptionLevel *int32
	Limit                   int32
	Offset                  int32
}

type SearchPostsQuery struct {
	Query                        string
	AuthorUserIDs                []int64
	SportTypeIDs                 []int64
	BlockKinds                   []domain.BlockKind
	MinRequiredSubscriptionLevel *int32
	MaxRequiredSubscriptionLevel *int32
	OnlyAvailable                bool
	ViewerUserID                 int64
	ViewerSubscriptionLevel      *int32
	Limit                        int32
	Offset                       int32
}

type CreatePostCommand struct {
	AuthorUserID              int64
	Title                     string
	RequiredSubscriptionLevel *int32
	SportTypeID               *int64
	Blocks                    []PostBlockInput
}

type UploadPostMediaCommand struct {
	AuthorUserID int64
	FileName     string
	ContentType  string
	Content      []byte
}

type GetPostQuery struct {
	PostID                  int64
	ViewerUserID            int64
	ViewerSubscriptionLevel *int32
}

type UpdatePostCommand struct {
	PostID                         int64
	AuthorUserID                   int64
	Title                          *string
	RequiredSubscriptionLevel      *int32
	ClearRequiredSubscriptionLevel bool
	SportTypeID                    *int64
	ClearSportTypeID               bool
	Blocks                         []PostBlockInput
	ReplaceBlocks                  bool
}

type ListSubscriptionTiersQuery struct {
	TrainerUserID int64
}

type CreateSubscriptionTierCommand struct {
	TrainerUserID int64
	Name          string
	Price         int32
	Description   *string
}

type UpdateSubscriptionTierCommand struct {
	TrainerUserID    int64
	TierID           int64
	Name             *string
	Price            *int32
	Description      *string
	ClearDescription bool
}

type DeleteSubscriptionTierCommand struct {
	TrainerUserID int64
	TierID        int64
}

type SubscribeToTrainerCommand struct {
	ClientUserID  int64
	TrainerUserID int64
	TierID        int64
}

type ListMySubscriptionsQuery struct {
	ClientUserID int64
}

type CancelSubscriptionCommand struct {
	ClientUserID   int64
	SubscriptionID int64
}

type DeletePostCommand struct {
	PostID       int64
	AuthorUserID int64
}

type LikePostCommand struct {
	PostID                  int64
	UserID                  int64
	ViewerSubscriptionLevel *int32
}

type CreateCommentCommand struct {
	PostID                  int64
	AuthorUserID            int64
	ViewerSubscriptionLevel *int32
	Body                    string
}

type ListCommentsQuery struct {
	PostID                  int64
	ViewerUserID            int64
	ViewerSubscriptionLevel *int32
	Limit                   int32
	Offset                  int32
}
