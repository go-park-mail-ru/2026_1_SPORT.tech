package usecase

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
)

type stubContentRepository struct {
	createPostFunc         func(ctx context.Context, post domain.Post) (int64, error)
	getPostFunc            func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error)
	listAuthorPostsFunc    func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error)
	searchPostsFunc        func(ctx context.Context, query SearchPostsQuery) ([]domain.PostSummary, error)
	updatePostFunc         func(ctx context.Context, post domain.Post, replaceBlocks bool) error
	deletePostFunc         func(ctx context.Context, postID int64, authorUserID int64) error
	listTiersFunc          func(ctx context.Context, trainerUserID int64) ([]domain.SubscriptionTier, error)
	getTierFunc            func(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error)
	createTierFunc         func(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error)
	updateTierFunc         func(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error)
	deleteTierFunc         func(ctx context.Context, trainerUserID int64, tierID int64) error
	activeLevelFunc        func(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error)
	subscribeFunc          func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	listSubscriptionsFunc  func(ctx context.Context, clientUserID int64) ([]domain.Subscription, error)
	updateSubscriptionFunc func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	cancelSubscriptionFunc func(ctx context.Context, clientUserID int64, subscriptionID int64) error
	upsertLikeFunc         func(ctx context.Context, postID int64, userID int64) error
	deleteLikeFunc         func(ctx context.Context, postID int64, userID int64) error
	getLikeStateFunc       func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error)
	createCommentFunc      func(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	listCommentsFunc       func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error)
	createDonationFunc     func(ctx context.Context, donation domain.Donation) (domain.Donation, error)
	getBalanceFunc         func(ctx context.Context, trainerUserID int64, currency string) (domain.Balance, error)
}

func (repository stubContentRepository) CreatePost(ctx context.Context, post domain.Post) (int64, error) {
	return repository.createPostFunc(ctx, post)
}

func (repository stubContentRepository) GetPost(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
	return repository.getPostFunc(ctx, postID, viewerUserID)
}

func (repository stubContentRepository) ListAuthorPosts(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
	return repository.listAuthorPostsFunc(ctx, authorUserID, viewerUserID, limit, offset)
}

func (repository stubContentRepository) SearchPosts(ctx context.Context, query SearchPostsQuery) ([]domain.PostSummary, error) {
	return repository.searchPostsFunc(ctx, query)
}

func (repository stubContentRepository) UpdatePost(ctx context.Context, post domain.Post, replaceBlocks bool) error {
	return repository.updatePostFunc(ctx, post, replaceBlocks)
}

func (repository stubContentRepository) DeletePost(ctx context.Context, postID int64, authorUserID int64) error {
	return repository.deletePostFunc(ctx, postID, authorUserID)
}

func (repository stubContentRepository) ListSubscriptionTiers(ctx context.Context, trainerUserID int64) ([]domain.SubscriptionTier, error) {
	if repository.listTiersFunc == nil {
		return nil, nil
	}
	return repository.listTiersFunc(ctx, trainerUserID)
}

func (repository stubContentRepository) GetSubscriptionTier(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error) {
	if repository.getTierFunc == nil {
		return domain.SubscriptionTier{}, nil
	}
	return repository.getTierFunc(ctx, trainerUserID, tierID)
}

func (repository stubContentRepository) CreateSubscriptionTier(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error) {
	if repository.createTierFunc == nil {
		return tier, nil
	}
	return repository.createTierFunc(ctx, tier)
}

func (repository stubContentRepository) UpdateSubscriptionTier(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error) {
	if repository.updateTierFunc == nil {
		return tier, nil
	}
	return repository.updateTierFunc(ctx, tier)
}

func (repository stubContentRepository) DeleteSubscriptionTier(ctx context.Context, trainerUserID int64, tierID int64) error {
	if repository.deleteTierFunc == nil {
		return nil
	}
	return repository.deleteTierFunc(ctx, trainerUserID, tierID)
}

func (repository stubContentRepository) GetActiveSubscriptionLevel(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error) {
	if repository.activeLevelFunc == nil {
		return nil, nil
	}
	return repository.activeLevelFunc(ctx, clientUserID, trainerUserID)
}

func (repository stubContentRepository) SubscribeToTrainer(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	if repository.subscribeFunc == nil {
		return subscription, nil
	}
	return repository.subscribeFunc(ctx, subscription)
}

func (repository stubContentRepository) ListSubscriptions(ctx context.Context, clientUserID int64) ([]domain.Subscription, error) {
	if repository.listSubscriptionsFunc == nil {
		return nil, nil
	}
	return repository.listSubscriptionsFunc(ctx, clientUserID)
}

func (repository stubContentRepository) UpdateSubscription(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
	if repository.updateSubscriptionFunc == nil {
		return subscription, nil
	}
	return repository.updateSubscriptionFunc(ctx, subscription)
}

func (repository stubContentRepository) CancelSubscription(ctx context.Context, clientUserID int64, subscriptionID int64) error {
	if repository.cancelSubscriptionFunc == nil {
		return nil
	}
	return repository.cancelSubscriptionFunc(ctx, clientUserID, subscriptionID)
}

func (repository stubContentRepository) UpsertLike(ctx context.Context, postID int64, userID int64) error {
	return repository.upsertLikeFunc(ctx, postID, userID)
}

func (repository stubContentRepository) DeleteLike(ctx context.Context, postID int64, userID int64) error {
	return repository.deleteLikeFunc(ctx, postID, userID)
}

func (repository stubContentRepository) GetPostLikeState(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
	return repository.getLikeStateFunc(ctx, postID, userID)
}

func (repository stubContentRepository) CreateComment(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	return repository.createCommentFunc(ctx, comment)
}

func (repository stubContentRepository) ListComments(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
	return repository.listCommentsFunc(ctx, postID, limit, offset)
}

func (repository stubContentRepository) CreateDonation(ctx context.Context, donation domain.Donation) (domain.Donation, error) {
	if repository.createDonationFunc == nil {
		return donation, nil
	}
	return repository.createDonationFunc(ctx, donation)
}

func (repository stubContentRepository) GetBalance(ctx context.Context, trainerUserID int64, currency string) (domain.Balance, error) {
	if repository.getBalanceFunc == nil {
		return domain.Balance{TrainerUserID: trainerUserID, Currency: currency}, nil
	}
	return repository.getBalanceFunc(ctx, trainerUserID, currency)
}

type stubPostMediaStorage struct {
	uploadFunc func(ctx context.Context, authorUserID int64, fileName string, contentType string, file io.Reader, size int64) (string, error)
}

func (storage stubPostMediaStorage) UploadPostMedia(
	ctx context.Context,
	authorUserID int64,
	fileName string,
	contentType string,
	file io.Reader,
	size int64,
) (string, error) {
	return storage.uploadFunc(ctx, authorUserID, fileName, contentType, file, size)
}

func TestServiceCreatePost(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)
	requiredLevel := int32(2)
	sportTypeID := int64(3001)

	service := NewService(
		stubContentRepository{
			createPostFunc: func(ctx context.Context, post domain.Post) (int64, error) {
				if post.AuthorUserID != 7 ||
					post.Title != "Morning run" ||
					post.SportTypeID == nil ||
					*post.SportTypeID != 3001 ||
					len(post.Blocks) != 2 {
					t.Fatalf("unexpected post: %+v", post)
				}
				return 101, nil
			},
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{
					PostID:                    postID,
					AuthorUserID:              viewerUserID,
					Title:                     "Morning run",
					RequiredSubscriptionLevel: &requiredLevel,
					SportTypeID:               &sportTypeID,
					CreatedAt:                 now,
					UpdatedAt:                 now,
					Blocks: []domain.PostBlock{{
						PostBlockID: 1,
						Position:    0,
						Kind:        domain.BlockKindText,
						TextContent: stringPtr("Warm-up"),
					}},
				}, nil
			},
			listAuthorPostsFunc: func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
				return nil, nil
			},
			updatePostFunc: func(ctx context.Context, post domain.Post, replaceBlocks bool) error { return nil },
			deletePostFunc: func(ctx context.Context, postID int64, authorUserID int64) error { return nil },
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			deleteLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			getLikeStateFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
				return domain.PostLikeState{}, nil
			},
			createCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
				return domain.Comment{}, nil
			},
			listCommentsFunc: func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
				return nil, nil
			},
		},
		nil,
	)

	post, err := service.CreatePost(context.Background(), CreatePostCommand{
		AuthorUserID:              7,
		Title:                     " Morning run ",
		RequiredSubscriptionLevel: &requiredLevel,
		SportTypeID:               &sportTypeID,
		Blocks: []PostBlockInput{
			{Kind: domain.BlockKindText, TextContent: stringPtr(" Warm-up ")},
			{Kind: domain.BlockKindImage, FileURL: stringPtr(" https://cdn.example/run.jpg ")},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post.PostID != 101 {
		t.Fatalf("unexpected post id: %d", post.PostID)
	}
}

func TestServiceUploadPostMedia(t *testing.T) {
	uploaded := false
	service := NewService(
		stubContentRepository{},
		stubPostMediaStorage{
			uploadFunc: func(ctx context.Context, authorUserID int64, fileName string, contentType string, file io.Reader, size int64) (string, error) {
				uploaded = true
				if authorUserID != 7 || fileName != "run.png" || contentType != "image/png" || size != 4 {
					t.Fatalf("unexpected upload args: authorUserID=%d fileName=%s contentType=%s size=%d", authorUserID, fileName, contentType, size)
				}

				content, err := io.ReadAll(file)
				if err != nil {
					t.Fatalf("read upload content: %v", err)
				}
				if string(content) != "data" {
					t.Fatalf("unexpected upload content: %q", string(content))
				}

				return "http://storage/posts/7/run.png", nil
			},
		},
	)

	media, err := service.UploadPostMedia(context.Background(), UploadPostMediaCommand{
		AuthorUserID: 7,
		FileName:     " run.png ",
		ContentType:  " image/png ",
		Content:      []byte("data"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !uploaded {
		t.Fatal("expected upload to be called")
	}
	if media.FileURL != "http://storage/posts/7/run.png" ||
		media.Kind != domain.BlockKindImage ||
		media.ContentType != "image/png" ||
		media.SizeBytes != 4 {
		t.Fatalf("unexpected media: %+v", media)
	}
}

func TestServiceSearchPostsAppliesFiltersAndAccessFlags(t *testing.T) {
	requiredLevel := int32(2)
	repositoryCalled := false
	service := NewService(
		stubContentRepository{
			searchPostsFunc: func(ctx context.Context, query SearchPostsQuery) ([]domain.PostSummary, error) {
				repositoryCalled = true
				if query.Query != "темп" ||
					len(query.AuthorUserIDs) != 1 ||
					query.AuthorUserIDs[0] != 7 ||
					len(query.SportTypeIDs) != 1 ||
					query.SportTypeIDs[0] != 3001 ||
					len(query.BlockKinds) != 1 ||
					query.BlockKinds[0] != domain.BlockKindImage ||
					query.Limit != 20 ||
					query.Offset != 10 ||
					query.ViewerUserID != 13 ||
					query.ViewerSubscriptionLevel == nil ||
					*query.ViewerSubscriptionLevel != 2 ||
					!query.OnlyAvailable {
					t.Fatalf("unexpected search query: %+v", query)
				}

				return []domain.PostSummary{{
					PostID:                    101,
					AuthorUserID:              7,
					Title:                     "Темповая тренировка",
					RequiredSubscriptionLevel: &requiredLevel,
				}}, nil
			},
			activeLevelFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error) {
				if clientUserID != 13 || trainerUserID != 7 {
					t.Fatalf("unexpected active subscription lookup: client=%d trainer=%d", clientUserID, trainerUserID)
				}
				return &requiredLevel, nil
			},
		},
		nil,
	)

	posts, err := service.SearchPosts(context.Background(), SearchPostsQuery{
		Query:                   " темп ",
		AuthorUserIDs:           []int64{7},
		SportTypeIDs:            []int64{3001},
		BlockKinds:              []domain.BlockKind{domain.BlockKindImage},
		OnlyAvailable:           true,
		ViewerUserID:            13,
		ViewerSubscriptionLevel: &requiredLevel,
		Limit:                   20,
		Offset:                  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repositoryCalled {
		t.Fatal("expected repository search to be called")
	}
	if len(posts) != 1 || !posts[0].CanView {
		t.Fatalf("unexpected posts: %+v", posts)
	}
}

func TestServiceCreateSubscriptionTier(t *testing.T) {
	service := NewService(
		stubContentRepository{
			createTierFunc: func(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error) {
				if tier.TrainerUserID != 7 ||
					tier.Name != "Продвинутый" ||
					tier.Price != 1500 ||
					tier.Description == nil ||
					*tier.Description != "Закрытые тренировки" {
					t.Fatalf("unexpected tier: %+v", tier)
				}

				tier.TierID = 2
				return tier, nil
			},
		},
		nil,
	)

	tier, err := service.CreateSubscriptionTier(context.Background(), CreateSubscriptionTierCommand{
		TrainerUserID: 7,
		Name:          " Продвинутый ",
		Price:         1500,
		Description:   stringPtr(" Закрытые тренировки "),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier.TierID != 2 {
		t.Fatalf("unexpected tier id: %d", tier.TierID)
	}
}

func TestServiceSubscribeToTrainer(t *testing.T) {
	service := NewService(
		stubContentRepository{
			getTierFunc: func(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error) {
				if trainerUserID != 1001 || tierID != 2 {
					t.Fatalf("unexpected tier lookup: trainer=%d tier=%d", trainerUserID, tierID)
				}

				return domain.SubscriptionTier{
					TrainerUserID: trainerUserID,
					TierID:        tierID,
					Name:          "Продвинутый",
					Price:         1500,
				}, nil
			},
			subscribeFunc: func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
				if subscription.ClientUserID != 1002 ||
					subscription.TrainerUserID != 1001 ||
					subscription.TierID != 2 ||
					subscription.ExpiresAt.IsZero() {
					t.Fatalf("unexpected subscription: %+v", subscription)
				}

				subscription.SubscriptionID = 2401
				subscription.TierName = "Продвинутый"
				subscription.Price = 1500
				subscription.Active = true
				return subscription, nil
			},
		},
		nil,
	)

	subscription, err := service.SubscribeToTrainer(context.Background(), SubscribeToTrainerCommand{
		ClientUserID:  1002,
		TrainerUserID: 1001,
		TierID:        2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subscription.SubscriptionID != 2401 || !subscription.Active {
		t.Fatalf("unexpected subscription result: %+v", subscription)
	}
}

func TestServiceUpdateSubscription(t *testing.T) {
	service := NewService(
		stubContentRepository{
			updateSubscriptionFunc: func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
				if subscription.ClientUserID != 1002 ||
					subscription.SubscriptionID != 2401 ||
					subscription.TierID != 3 {
					t.Fatalf("unexpected subscription update: %+v", subscription)
				}

				subscription.TrainerUserID = 1001
				subscription.TierName = "Премиум"
				subscription.Price = 2500
				subscription.Active = true
				return subscription, nil
			},
		},
		nil,
	)

	subscription, err := service.UpdateSubscription(context.Background(), UpdateSubscriptionCommand{
		ClientUserID:   1002,
		SubscriptionID: 2401,
		TierID:         3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subscription.TierID != 3 || subscription.TierName != "Премиум" || !subscription.Active {
		t.Fatalf("unexpected subscription result: %+v", subscription)
	}
}

func TestServiceGetPostRejectsRestrictedAccess(t *testing.T) {
	requiredLevel := int32(2)

	service := NewService(
		stubContentRepository{
			createPostFunc: func(ctx context.Context, post domain.Post) (int64, error) { return 0, nil },
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{
					PostID:                    postID,
					AuthorUserID:              9,
					Title:                     "Subscribers only",
					RequiredSubscriptionLevel: &requiredLevel,
				}, nil
			},
			listAuthorPostsFunc: func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
				return nil, nil
			},
			updatePostFunc: func(ctx context.Context, post domain.Post, replaceBlocks bool) error { return nil },
			deletePostFunc: func(ctx context.Context, postID int64, authorUserID int64) error { return nil },
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			deleteLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			getLikeStateFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
				return domain.PostLikeState{}, nil
			},
			createCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
				return domain.Comment{}, nil
			},
			listCommentsFunc: func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
				return nil, nil
			},
		},
		nil,
	)

	_, err := service.GetPost(context.Background(), GetPostQuery{
		PostID:       33,
		ViewerUserID: 7,
	})
	if !errors.Is(err, domain.ErrPostForbidden) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreateComment(t *testing.T) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)

	service := NewService(
		stubContentRepository{
			createPostFunc: func(ctx context.Context, post domain.Post) (int64, error) { return 0, nil },
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{
					PostID:       postID,
					AuthorUserID: 7,
					Title:        "Public post",
				}, nil
			},
			listAuthorPostsFunc: func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
				return nil, nil
			},
			updatePostFunc: func(ctx context.Context, post domain.Post, replaceBlocks bool) error { return nil },
			deletePostFunc: func(ctx context.Context, postID int64, authorUserID int64) error { return nil },
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			deleteLikeFunc: func(ctx context.Context, postID int64, userID int64) error { return nil },
			getLikeStateFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
				return domain.PostLikeState{}, nil
			},
			createCommentFunc: func(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
				if comment.PostID != 21 || comment.AuthorUserID != 13 || comment.Body != "Great workout" {
					t.Fatalf("unexpected comment: %+v", comment)
				}
				comment.CommentID = 88
				comment.CreatedAt = now
				comment.UpdatedAt = now
				return comment, nil
			},
			listCommentsFunc: func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
				return nil, nil
			},
		},
		nil,
	)

	comment, err := service.CreateComment(context.Background(), CreateCommentCommand{
		PostID:       21,
		AuthorUserID: 13,
		Body:         " Great workout ",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.CommentID != 88 {
		t.Fatalf("unexpected comment id: %d", comment.CommentID)
	}
}

func stringPtr(value string) *string {
	return &value
}
