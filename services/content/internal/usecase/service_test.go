package usecase

import (
	"context"
	"errors"
	"io"
	"strings"
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
	listPriceIncreaseFunc  func(ctx context.Context, trainerUserID int64, tierID int64, newPrice int32) ([]domain.Subscription, error)
	blockRenewalFunc       func(ctx context.Context, subscriptionID int64) error
	activeLevelFunc        func(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error)
	subscribeFunc          func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	listSubscriptionsFunc  func(ctx context.Context, clientUserID int64) ([]domain.Subscription, error)
	listSubscribersFunc    func(ctx context.Context, trainerUserID int64, limit int32, offset int32) ([]domain.Subscription, error)
	updateSubscriptionFunc func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error)
	cancelSubscriptionFunc func(ctx context.Context, clientUserID int64, subscriptionID int64) error
	upsertLikeFunc         func(ctx context.Context, postID int64, userID int64) (bool, error)
	deleteLikeFunc         func(ctx context.Context, postID int64, userID int64) error
	getLikeStateFunc       func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error)
	createCommentFunc      func(ctx context.Context, comment domain.Comment) (domain.Comment, error)
	listCommentsFunc       func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error)
	getCommentFunc         func(ctx context.Context, commentID int64) (domain.Comment, error)
	updateCommentFunc      func(ctx context.Context, commentID int64, body string, now time.Time) (domain.Comment, error)
	deleteCommentFunc      func(ctx context.Context, commentID int64) error
	listPostLikesFunc      func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.PostLike, error)
	createDonationFunc     func(ctx context.Context, donation domain.Donation) (domain.Donation, error)
	createPaymentFunc      func(ctx context.Context, payment domain.DonationPayment) (domain.DonationPayment, error)
	confirmPaymentFunc     func(ctx context.Context, senderUserID int64, paymentID int64, confirmationToken string) (domain.DonationPayment, bool, error)
	confirmByProviderFunc  func(ctx context.Context, providerPaymentID string) (domain.DonationPayment, bool, error)
	setPaymentStatusFunc   func(ctx context.Context, providerPaymentID string, status domain.PaymentStatus) (domain.DonationPayment, bool, error)
	listStalePaymentsFunc  func(ctx context.Context, olderThan time.Time, limit int32) ([]domain.DonationPayment, error)
	deactivateExpiredFunc  func(ctx context.Context, expiredBefore time.Time) (int64, error)
	getSubscriptionFunc    func(ctx context.Context, clientUserID int64, subscriptionID int64) (domain.Subscription, error)
	setAutoRenewFunc       func(ctx context.Context, clientUserID int64, subscriptionID int64, autoRenew bool) error
	renewByStripeFunc      func(ctx context.Context, stripeSubscriptionID string, currentPeriodEnd time.Time) (bool, error)
	deactivateByStripeFunc func(ctx context.Context, stripeSubscriptionID string) (bool, error)
	getBalanceFunc         func(ctx context.Context, trainerUserID int64, currency string) (domain.Balance, error)
	getStatisticsFunc      func(ctx context.Context, trainerUserID int64, currency string, monthStart time.Time) (domain.TrainerStatistics, error)
	createNotificationFunc func(ctx context.Context, notification domain.Notification) (domain.Notification, error)
	listNotificationsFunc  func(ctx context.Context, userID int64, limit int32, offset int32) ([]domain.Notification, error)
	markNotificationFunc   func(ctx context.Context, userID int64, notificationID int64) (domain.Notification, error)
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

func (repository stubContentRepository) ListSubscriptionsAffectedByTierPriceIncrease(ctx context.Context, trainerUserID int64, tierID int64, newPrice int32) ([]domain.Subscription, error) {
	if repository.listPriceIncreaseFunc == nil {
		return nil, nil
	}
	return repository.listPriceIncreaseFunc(ctx, trainerUserID, tierID, newPrice)
}

func (repository stubContentRepository) BlockSubscriptionRenewalForPriceIncrease(ctx context.Context, subscriptionID int64) error {
	if repository.blockRenewalFunc == nil {
		return nil
	}
	return repository.blockRenewalFunc(ctx, subscriptionID)
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

func (repository stubContentRepository) ListTrainerSubscribers(ctx context.Context, trainerUserID int64, limit int32, offset int32) ([]domain.Subscription, error) {
	if repository.listSubscribersFunc == nil {
		return nil, nil
	}
	return repository.listSubscribersFunc(ctx, trainerUserID, limit, offset)
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

func (repository stubContentRepository) UpsertLike(ctx context.Context, postID int64, userID int64) (bool, error) {
	if repository.upsertLikeFunc == nil {
		return true, nil
	}
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

func (repository stubContentRepository) GetComment(ctx context.Context, commentID int64) (domain.Comment, error) {
	if repository.getCommentFunc == nil {
		return domain.Comment{}, nil
	}
	return repository.getCommentFunc(ctx, commentID)
}

func (repository stubContentRepository) UpdateComment(ctx context.Context, commentID int64, body string, now time.Time) (domain.Comment, error) {
	if repository.updateCommentFunc == nil {
		return domain.Comment{}, nil
	}
	return repository.updateCommentFunc(ctx, commentID, body, now)
}

func (repository stubContentRepository) DeleteComment(ctx context.Context, commentID int64) error {
	if repository.deleteCommentFunc == nil {
		return nil
	}
	return repository.deleteCommentFunc(ctx, commentID)
}

func (repository stubContentRepository) ListPostLikes(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.PostLike, error) {
	if repository.listPostLikesFunc == nil {
		return nil, nil
	}
	return repository.listPostLikesFunc(ctx, postID, limit, offset)
}

func (repository stubContentRepository) CreateDonation(ctx context.Context, donation domain.Donation) (domain.Donation, error) {
	if repository.createDonationFunc == nil {
		return donation, nil
	}
	return repository.createDonationFunc(ctx, donation)
}

func (repository stubContentRepository) CreateDonationPayment(ctx context.Context, payment domain.DonationPayment) (domain.DonationPayment, error) {
	if repository.createPaymentFunc == nil {
		return payment, nil
	}
	return repository.createPaymentFunc(ctx, payment)
}

func (repository stubContentRepository) GetDonationPayment(ctx context.Context, senderUserID int64, paymentID int64) (domain.DonationPayment, error) {
	return domain.DonationPayment{
		PaymentID:         paymentID,
		SenderUserID:      senderUserID,
		ProviderPaymentID: "provider-payment-1",
		ConfirmationToken: "confirm_abc",
		Status:            domain.PaymentStatusPending,
	}, nil
}

func (repository stubContentRepository) ConfirmDonationPayment(ctx context.Context, senderUserID int64, paymentID int64, confirmationToken string) (domain.DonationPayment, bool, error) {
	if repository.confirmPaymentFunc == nil {
		return domain.DonationPayment{PaymentID: paymentID, SenderUserID: senderUserID, ConfirmationToken: confirmationToken}, true, nil
	}
	return repository.confirmPaymentFunc(ctx, senderUserID, paymentID, confirmationToken)
}

func (repository stubContentRepository) ConfirmPaymentByProviderID(ctx context.Context, providerPaymentID string, stripeSubscriptionID string) (domain.DonationPayment, bool, error) {
	if repository.confirmByProviderFunc == nil {
		return domain.DonationPayment{ProviderPaymentID: providerPaymentID, Status: domain.PaymentStatusConfirmed}, true, nil
	}
	return repository.confirmByProviderFunc(ctx, providerPaymentID)
}

func (repository stubContentRepository) GetSubscription(ctx context.Context, clientUserID int64, subscriptionID int64) (domain.Subscription, error) {
	if repository.getSubscriptionFunc == nil {
		return domain.Subscription{SubscriptionID: subscriptionID, ClientUserID: clientUserID}, nil
	}
	return repository.getSubscriptionFunc(ctx, clientUserID, subscriptionID)
}

func (repository stubContentRepository) SetSubscriptionAutoRenew(ctx context.Context, clientUserID int64, subscriptionID int64, autoRenew bool) error {
	if repository.setAutoRenewFunc == nil {
		return nil
	}
	return repository.setAutoRenewFunc(ctx, clientUserID, subscriptionID, autoRenew)
}

func (repository stubContentRepository) RenewSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string, currentPeriodEnd time.Time) (bool, error) {
	if repository.renewByStripeFunc == nil {
		return true, nil
	}
	return repository.renewByStripeFunc(ctx, stripeSubscriptionID, currentPeriodEnd)
}

func (repository stubContentRepository) DeactivateSubscriptionByStripeID(ctx context.Context, stripeSubscriptionID string) (bool, error) {
	if repository.deactivateByStripeFunc == nil {
		return true, nil
	}
	return repository.deactivateByStripeFunc(ctx, stripeSubscriptionID)
}

func (repository stubContentRepository) SetPaymentStatusByProviderID(ctx context.Context, providerPaymentID string, status domain.PaymentStatus) (domain.DonationPayment, bool, error) {
	if repository.setPaymentStatusFunc == nil {
		return domain.DonationPayment{ProviderPaymentID: providerPaymentID, Status: status}, true, nil
	}
	return repository.setPaymentStatusFunc(ctx, providerPaymentID, status)
}

func (repository stubContentRepository) ListStalePendingPayments(ctx context.Context, olderThan time.Time, limit int32) ([]domain.DonationPayment, error) {
	if repository.listStalePaymentsFunc == nil {
		return nil, nil
	}
	return repository.listStalePaymentsFunc(ctx, olderThan, limit)
}

func (repository stubContentRepository) DeactivateExpiredSubscriptions(ctx context.Context, expiredBefore time.Time) (int64, error) {
	if repository.deactivateExpiredFunc == nil {
		return 0, nil
	}
	return repository.deactivateExpiredFunc(ctx, expiredBefore)
}

func (repository stubContentRepository) GetBalance(ctx context.Context, trainerUserID int64, currency string) (domain.Balance, error) {
	if repository.getBalanceFunc == nil {
		return domain.Balance{TrainerUserID: trainerUserID, Currency: currency}, nil
	}
	return repository.getBalanceFunc(ctx, trainerUserID, currency)
}

func (repository stubContentRepository) GetTrainerStatistics(ctx context.Context, trainerUserID int64, currency string, monthStart time.Time) (domain.TrainerStatistics, error) {
	if repository.getStatisticsFunc == nil {
		return domain.TrainerStatistics{TrainerUserID: trainerUserID, Currency: currency}, nil
	}
	return repository.getStatisticsFunc(ctx, trainerUserID, currency, monthStart)
}

func (repository stubContentRepository) ListReceivedDonations(ctx context.Context, recipientUserID int64, limit, offset int32) ([]domain.Donation, error) {
	return nil, nil
}

func (repository stubContentRepository) CountReceivedDonations(ctx context.Context, recipientUserID int64) (int32, error) {
	return 0, nil
}

func (repository stubContentRepository) CreateNotification(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
	if repository.createNotificationFunc == nil {
		return notification, nil
	}
	return repository.createNotificationFunc(ctx, notification)
}

func (repository stubContentRepository) ListNotifications(ctx context.Context, userID int64, limit int32, offset int32) ([]domain.Notification, error) {
	if repository.listNotificationsFunc == nil {
		return nil, nil
	}
	return repository.listNotificationsFunc(ctx, userID, limit, offset)
}

func (repository stubContentRepository) MarkNotificationRead(ctx context.Context, userID int64, notificationID int64) (domain.Notification, error) {
	if repository.markNotificationFunc == nil {
		return domain.Notification{NotificationID: notificationID, UserID: userID}, nil
	}
	return repository.markNotificationFunc(ctx, userID, notificationID)
}

type stubPaymentProvider struct {
	createFunc func(ctx context.Context, request PaymentProviderCreateRequest) (PaymentProviderPayment, error)
	getFunc    func(ctx context.Context, providerPaymentID string) (PaymentProviderPayment, error)
	cancelFunc func(ctx context.Context, providerSubscriptionID string, atPeriodEnd bool) error
}

func (provider stubPaymentProvider) CancelSubscription(ctx context.Context, providerSubscriptionID string, atPeriodEnd bool) error {
	if provider.cancelFunc == nil {
		return nil
	}
	return provider.cancelFunc(ctx, providerSubscriptionID, atPeriodEnd)
}

func (provider stubPaymentProvider) CreatePayment(ctx context.Context, request PaymentProviderCreateRequest) (PaymentProviderPayment, error) {
	if provider.createFunc == nil {
		return PaymentProviderPayment{ProviderPaymentID: "provider-payment-1", Status: "pending", ConfirmationURL: "https://pay.example/1"}, nil
	}
	return provider.createFunc(ctx, request)
}

func (provider stubPaymentProvider) ProviderName() string {
	return "stripe"
}

func (provider stubPaymentProvider) GetPayment(ctx context.Context, providerPaymentID string) (PaymentProviderPayment, error) {
	if provider.getFunc == nil {
		return PaymentProviderPayment{ProviderPaymentID: providerPaymentID, Status: "succeeded"}, nil
	}
	return provider.getFunc(ctx, providerPaymentID)
}

func stubRepositories(repository stubContentRepository) Repositories {
	return Repositories{
		Posts:         repository,
		Money:         repository,
		Engagement:    repository,
		Notifications: repository,
	}
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
		stubRepositories(stubContentRepository{
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
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) (bool, error) { return true, nil },
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
		}),
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

func TestServiceCreatePostNotifiesEligibleSubscribers(t *testing.T) {
	requiredLevel := int32(2)
	var notifications []domain.Notification
	service := NewService(
		stubRepositories(stubContentRepository{
			createPostFunc: func(ctx context.Context, post domain.Post) (int64, error) {
				return 55, nil
			},
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{
					PostID:                    postID,
					AuthorUserID:              viewerUserID,
					Title:                     "Paid workout",
					RequiredSubscriptionLevel: &requiredLevel,
				}, nil
			},
			listSubscribersFunc: func(ctx context.Context, trainerUserID int64, limit int32, offset int32) ([]domain.Subscription, error) {
				if trainerUserID != 7 || limit != maxPageLimit || offset != 0 {
					t.Fatalf("unexpected subscribers query: trainer=%d limit=%d offset=%d", trainerUserID, limit, offset)
				}
				return []domain.Subscription{
					{ClientUserID: 1001, TrainerUserID: trainerUserID, TierID: 1},
					{ClientUserID: 1002, TrainerUserID: trainerUserID, TierID: 2},
				}, nil
			},
			createNotificationFunc: func(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
				notifications = append(notifications, notification)
				return notification, nil
			},
		}),
		nil,
	)

	if _, err := service.CreatePost(context.Background(), CreatePostCommand{
		AuthorUserID:              7,
		Title:                     "Paid workout",
		RequiredSubscriptionLevel: &requiredLevel,
		Blocks: []PostBlockInput{
			{Kind: domain.BlockKindText, TextContent: stringPtr("Workout plan")},
		},
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("unexpected notifications: %+v", notifications)
	}
	if notifications[0].UserID != 1002 ||
		notifications[0].ActorUserID != 7 ||
		notifications[0].Type != domain.NotificationTypePost ||
		notifications[0].PostID == nil ||
		*notifications[0].PostID != 55 {
		t.Fatalf("unexpected notification: %+v", notifications[0])
	}
}

func TestServiceUploadPostMedia(t *testing.T) {
	uploaded := false
	service := NewService(
		stubRepositories(stubContentRepository{}),
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
		stubRepositories(stubContentRepository{
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
		}),
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
		stubRepositories(stubContentRepository{
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
		}),
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
		stubRepositories(stubContentRepository{
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
		}),
		nil,
	)

	_, err := service.SubscribeToTrainer(context.Background(), SubscribeToTrainerCommand{
		ClientUserID:  1002,
		TrainerUserID: 1001,
		TierID:        2,
	})
	if !errors.Is(err, ErrSubscriptionPaymentRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceUpdateSubscription(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
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
		}),
		nil,
	)

	_, err := service.UpdateSubscription(context.Background(), UpdateSubscriptionCommand{
		ClientUserID:   1002,
		SubscriptionID: 2401,
		TierID:         3,
	})
	if !errors.Is(err, ErrSubscriptionPaymentRequired) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceUpdateSubscriptionTier(t *testing.T) {
	description := "Новое описание"
	price := int32(900)
	service := NewService(
		stubRepositories(stubContentRepository{
			getTierFunc: func(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error) {
				if trainerUserID != 7 || tierID != 2 {
					t.Fatalf("unexpected tier lookup: trainer=%d tier=%d", trainerUserID, tierID)
				}
				return domain.SubscriptionTier{
					TrainerUserID: trainerUserID,
					TierID:        tierID,
					Name:          "Базовый",
					Price:         500,
					Description:   stringPtr("Старое описание"),
				}, nil
			},
			updateTierFunc: func(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error) {
				if tier.Name != "Премиум" || tier.Price != 900 || tier.Description == nil || *tier.Description != "Новое описание" {
					t.Fatalf("unexpected tier update: %+v", tier)
				}
				return tier, nil
			},
		}),
		nil,
	)

	tier, err := service.UpdateSubscriptionTier(context.Background(), UpdateSubscriptionTierCommand{
		TrainerUserID: 7,
		TierID:        2,
		Name:          stringPtr(" Премиум "),
		Price:         &price,
		Description:   &description,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier.Name != "Премиум" || tier.Price != 900 {
		t.Fatalf("unexpected tier result: %+v", tier)
	}
}

func TestServiceUpdateSubscriptionTierPriceIncreaseCancelsRenewals(t *testing.T) {
	price := int32(900)
	periodEnd := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	canceledProviderSubscriptions := make([]string, 0)
	blockedSubscriptions := make([]int64, 0)
	notifications := make([]domain.Notification, 0)

	service := NewService(
		stubRepositories(stubContentRepository{
			getTierFunc: func(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error) {
				return domain.SubscriptionTier{
					TrainerUserID: trainerUserID,
					TierID:        tierID,
					Name:          "Базовый",
					Price:         500,
				}, nil
			},
			listPriceIncreaseFunc: func(ctx context.Context, trainerUserID int64, tierID int64, newPrice int32) ([]domain.Subscription, error) {
				if trainerUserID != 7 || tierID != 2 || newPrice != 900 {
					t.Fatalf("unexpected price increase lookup: trainer=%d tier=%d price=%d", trainerUserID, tierID, newPrice)
				}
				return []domain.Subscription{{
					SubscriptionID:       11,
					ClientUserID:         1002,
					TrainerUserID:        trainerUserID,
					TierID:               tierID,
					TierName:             "Базовый",
					Price:                500,
					Active:               true,
					ExpiresAt:            periodEnd,
					StripeSubscriptionID: "sub_123",
					CurrentPeriodEnd:     &periodEnd,
					AutoRenew:            true,
				}}, nil
			},
			blockRenewalFunc: func(ctx context.Context, subscriptionID int64) error {
				blockedSubscriptions = append(blockedSubscriptions, subscriptionID)
				return nil
			},
			updateTierFunc: func(ctx context.Context, tier domain.SubscriptionTier) (domain.SubscriptionTier, error) {
				if tier.Price != 900 {
					t.Fatalf("unexpected tier update: %+v", tier)
				}
				return tier, nil
			},
			createNotificationFunc: func(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
				notifications = append(notifications, notification)
				return notification, nil
			},
		}),
		nil,
		stubPaymentProvider{
			cancelFunc: func(ctx context.Context, providerSubscriptionID string, atPeriodEnd bool) error {
				if !atPeriodEnd {
					t.Fatal("expected cancellation at period end")
				}
				canceledProviderSubscriptions = append(canceledProviderSubscriptions, providerSubscriptionID)
				return nil
			},
		},
	)

	if _, err := service.UpdateSubscriptionTier(context.Background(), UpdateSubscriptionTierCommand{
		TrainerUserID: 7,
		TierID:        2,
		Price:         &price,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(canceledProviderSubscriptions) != 1 || canceledProviderSubscriptions[0] != "sub_123" {
		t.Fatalf("unexpected provider cancellations: %+v", canceledProviderSubscriptions)
	}
	if len(blockedSubscriptions) != 1 || blockedSubscriptions[0] != 11 {
		t.Fatalf("unexpected blocked subscriptions: %+v", blockedSubscriptions)
	}
	if len(notifications) != 1 ||
		notifications[0].UserID != 1002 ||
		notifications[0].ActorUserID != 7 ||
		notifications[0].Type != domain.NotificationTypeSubscription {
		t.Fatalf("unexpected notifications: %+v", notifications)
	}
}

func TestServiceListAndDeleteSubscriptionTiers(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			listTiersFunc: func(ctx context.Context, trainerUserID int64) ([]domain.SubscriptionTier, error) {
				if trainerUserID != 7 {
					t.Fatalf("unexpected trainer id: %d", trainerUserID)
				}
				return []domain.SubscriptionTier{{TierID: 1, TrainerUserID: 7, Name: "Базовый"}}, nil
			},
			deleteTierFunc: func(ctx context.Context, trainerUserID int64, tierID int64) error {
				if trainerUserID != 7 || tierID != 1 {
					t.Fatalf("unexpected delete args: trainer=%d tier=%d", trainerUserID, tierID)
				}
				return nil
			},
		}),
		nil,
	)

	tiers, err := service.ListSubscriptionTiers(context.Background(), ListSubscriptionTiersQuery{TrainerUserID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tiers) != 1 {
		t.Fatalf("unexpected tiers: %+v", tiers)
	}
	if err := service.DeleteSubscriptionTier(context.Background(), DeleteSubscriptionTierCommand{TrainerUserID: 7, TierID: 1}); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
}

func TestServiceListAndCancelSubscriptions(t *testing.T) {
	periodEnd := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	notifications := make([]domain.Notification, 0, 2)
	service := NewService(
		stubRepositories(stubContentRepository{
			listSubscriptionsFunc: func(ctx context.Context, clientUserID int64) ([]domain.Subscription, error) {
				if clientUserID != 1002 {
					t.Fatalf("unexpected client id: %d", clientUserID)
				}
				return []domain.Subscription{{SubscriptionID: 1, ClientUserID: 1002}}, nil
			},
			listSubscribersFunc: func(ctx context.Context, trainerUserID int64, limit int32, offset int32) ([]domain.Subscription, error) {
				if trainerUserID != 1001 || limit != 20 || offset != 5 {
					t.Fatalf("unexpected subscribers args: trainer=%d limit=%d offset=%d", trainerUserID, limit, offset)
				}
				return []domain.Subscription{{SubscriptionID: 2, ClientUserID: 1002, TrainerUserID: trainerUserID}}, nil
			},
			cancelSubscriptionFunc: func(ctx context.Context, clientUserID int64, subscriptionID int64) error {
				if clientUserID != 1002 || subscriptionID != 1 {
					t.Fatalf("unexpected cancel args: client=%d subscription=%d", clientUserID, subscriptionID)
				}
				return nil
			},
			getSubscriptionFunc: func(ctx context.Context, clientUserID int64, subscriptionID int64) (domain.Subscription, error) {
				return domain.Subscription{
					SubscriptionID: subscriptionID,
					ClientUserID:   clientUserID,
					TrainerUserID:  1001,
					TierName:       "Продвинутый",
					Active:         true,
					ExpiresAt:      periodEnd,
				}, nil
			},
			createNotificationFunc: func(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
				notifications = append(notifications, notification)
				return notification, nil
			},
		}),
		nil,
	)

	subscriptions, err := service.ListMySubscriptions(context.Background(), ListMySubscriptionsQuery{ClientUserID: 1002})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subscriptions) != 1 {
		t.Fatalf("unexpected subscriptions: %+v", subscriptions)
	}
	subscribers, err := service.ListTrainerSubscribers(context.Background(), ListTrainerSubscribersQuery{
		TrainerUserID: 1001,
		Offset:        5,
	})
	if err != nil {
		t.Fatalf("unexpected subscribers error: %v", err)
	}
	if len(subscribers) != 1 || subscribers[0].ClientUserID != 1002 {
		t.Fatalf("unexpected subscribers: %+v", subscribers)
	}
	if err := service.CancelSubscription(context.Background(), CancelSubscriptionCommand{ClientUserID: 1002, SubscriptionID: 1}); err != nil {
		t.Fatalf("unexpected cancel error: %v", err)
	}
	if len(notifications) != 2 {
		t.Fatalf("unexpected cancellation notifications: %+v", notifications)
	}
	if notifications[0].UserID != 1002 ||
		notifications[0].ActorUserID != 1001 ||
		notifications[0].Type != domain.NotificationTypeSubscription ||
		notifications[0].Title != "Вы отписались" {
		t.Fatalf("unexpected client cancellation notification: %+v", notifications[0])
	}
	if notifications[1].UserID != 1001 ||
		notifications[1].ActorUserID != 1002 ||
		notifications[1].Type != domain.NotificationTypeSubscription ||
		notifications[1].Title != "Подписчик отписался" {
		t.Fatalf("unexpected trainer cancellation notification: %+v", notifications[1])
	}
}

func TestServiceCancelSubscriptionViaStripe(t *testing.T) {
	var canceledAtPeriodEnd *bool
	autoRenewSet := true
	localCancelCalled := false
	periodEnd := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	notifications := make([]domain.Notification, 0, 2)
	service := NewService(
		stubRepositories(stubContentRepository{
			getSubscriptionFunc: func(ctx context.Context, clientUserID int64, subscriptionID int64) (domain.Subscription, error) {
				return domain.Subscription{
					SubscriptionID:       subscriptionID,
					ClientUserID:         clientUserID,
					TrainerUserID:        1001,
					TierName:             "Премиум",
					Active:               true,
					ExpiresAt:            periodEnd,
					StripeSubscriptionID: "sub_123",
					CurrentPeriodEnd:     &periodEnd,
					AutoRenew:            true,
				}, nil
			},
			setAutoRenewFunc: func(ctx context.Context, clientUserID int64, subscriptionID int64, autoRenew bool) error {
				autoRenewSet = autoRenew
				return nil
			},
			cancelSubscriptionFunc: func(ctx context.Context, clientUserID int64, subscriptionID int64) error {
				localCancelCalled = true
				return nil
			},
			createNotificationFunc: func(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
				notifications = append(notifications, notification)
				return notification, nil
			},
		}),
		nil,
		stubPaymentProvider{
			cancelFunc: func(ctx context.Context, providerSubscriptionID string, atPeriodEnd bool) error {
				if providerSubscriptionID != "sub_123" {
					t.Fatalf("unexpected provider subscription id: %s", providerSubscriptionID)
				}
				canceledAtPeriodEnd = &atPeriodEnd
				return nil
			},
		},
	)

	if err := service.CancelSubscription(context.Background(), CancelSubscriptionCommand{ClientUserID: 1002, SubscriptionID: 7}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if canceledAtPeriodEnd == nil || !*canceledAtPeriodEnd {
		t.Fatalf("expected stripe cancel at period end, got %v", canceledAtPeriodEnd)
	}
	if autoRenewSet {
		t.Fatal("expected auto_renew to be set false")
	}
	if localCancelCalled {
		t.Fatal("expected no immediate local cancel for stripe subscription")
	}
	if len(notifications) != 2 ||
		notifications[0].Title != "Вы отписались" ||
		notifications[1].Title != "Подписчик отписался" {
		t.Fatalf("unexpected stripe cancellation notifications: %+v", notifications)
	}
}

func TestServiceCancelSubscriptionAlreadyCancelled(t *testing.T) {
	providerCalled := false
	notificationCalled := false
	service := NewService(
		stubRepositories(stubContentRepository{
			getSubscriptionFunc: func(ctx context.Context, clientUserID int64, subscriptionID int64) (domain.Subscription, error) {
				return domain.Subscription{
					SubscriptionID:       subscriptionID,
					ClientUserID:         clientUserID,
					Active:               true,
					StripeSubscriptionID: "sub_123",
					AutoRenew:            false,
				}, nil
			},
			createNotificationFunc: func(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
				notificationCalled = true
				return notification, nil
			},
		}),
		nil,
		stubPaymentProvider{
			cancelFunc: func(ctx context.Context, providerSubscriptionID string, atPeriodEnd bool) error {
				providerCalled = true
				return nil
			},
		},
	)

	if err := service.CancelSubscription(context.Background(), CancelSubscriptionCommand{ClientUserID: 1002, SubscriptionID: 7}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if providerCalled {
		t.Fatal("expected already cancelled subscription to skip provider cancel")
	}
	if notificationCalled {
		t.Fatal("expected already cancelled subscription to skip cancellation notifications")
	}
}

func TestServiceCancelSubscriptionNotFoundIsIdempotent(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			getSubscriptionFunc: func(ctx context.Context, clientUserID int64, subscriptionID int64) (domain.Subscription, error) {
				return domain.Subscription{}, domain.ErrSubscriptionNotFound
			},
		}),
		nil,
	)

	if err := service.CancelSubscription(context.Background(), CancelSubscriptionCommand{ClientUserID: 1002, SubscriptionID: 7}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceConfirmSubscriptionPaymentRejectsEmptyStripeID(t *testing.T) {
	service := NewService(stubRepositories(stubContentRepository{}), nil)

	if _, err := service.ConfirmSubscriptionPaymentFromProvider(context.Background(), "cs_1", " "); !errors.Is(err, ErrInvalidProviderSubscriptionID) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceRenewAndDeactivateSubscriptionFromProvider(t *testing.T) {
	renewed := false
	deactivated := false
	service := NewService(
		stubRepositories(stubContentRepository{
			renewByStripeFunc: func(ctx context.Context, stripeSubscriptionID string, currentPeriodEnd time.Time) (bool, error) {
				if stripeSubscriptionID != "sub_123" {
					t.Fatalf("unexpected renew id: %s", stripeSubscriptionID)
				}
				renewed = true
				return true, nil
			},
			deactivateByStripeFunc: func(ctx context.Context, stripeSubscriptionID string) (bool, error) {
				if stripeSubscriptionID != "sub_123" {
					t.Fatalf("unexpected deactivate id: %s", stripeSubscriptionID)
				}
				deactivated = true
				return true, nil
			},
		}),
		nil,
	)

	if err := service.RenewSubscriptionFromProvider(context.Background(), "sub_123", time.Now().UTC()); err != nil {
		t.Fatalf("unexpected renew error: %v", err)
	}
	if err := service.DeactivateSubscriptionFromProvider(context.Background(), "sub_123"); err != nil {
		t.Fatalf("unexpected deactivate error: %v", err)
	}
	if !renewed || !deactivated {
		t.Fatalf("expected renew and deactivate to be called: renewed=%v deactivated=%v", renewed, deactivated)
	}
}

func TestServiceDonateToProfile(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			createDonationFunc: func(ctx context.Context, donation domain.Donation) (domain.Donation, error) {
				if donation.SenderUserID != 1002 ||
					donation.RecipientUserID != 1001 ||
					donation.AmountValue != 1500 ||
					donation.Currency != "RUB" ||
					donation.Message == nil ||
					*donation.Message != "Спасибо за тренировку" {
					t.Fatalf("unexpected donation: %+v", donation)
				}

				donation.DonationID = 77
				return donation, nil
			},
		}),
		nil,
	)

	donation, err := service.DonateToProfile(context.Background(), DonateToProfileCommand{
		SenderUserID:    1002,
		RecipientUserID: 1001,
		AmountValue:     1500,
		Message:         stringPtr(" Спасибо за тренировку "),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if donation.DonationID != 77 || donation.Currency != "RUB" {
		t.Fatalf("unexpected donation result: %+v", donation)
	}
}

func TestServiceDonateToProfileRejectsTooLargeAmount(t *testing.T) {
	service := NewService(stubRepositories(stubContentRepository{}), nil)

	_, err := service.DonateToProfile(context.Background(), DonateToProfileCommand{
		SenderUserID:    1002,
		RecipientUserID: 1001,
		AmountValue:     maxDonationAmount + 1,
	})
	if !errors.Is(err, ErrInvalidDonationAmount) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceCreateDonationPaymentRejectsTooSmallAmount(t *testing.T) {
	service := NewService(stubRepositories(stubContentRepository{}), nil)

	_, err := service.CreateDonationPayment(context.Background(), CreateDonationPaymentCommand{
		SenderUserID:    1002,
		RecipientUserID: 1001,
		AmountValue:     minDonationAmount - 1,
	})
	if !errors.Is(err, ErrInvalidDonationAmount) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceDonationPaymentFlow(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			createPaymentFunc: func(ctx context.Context, payment domain.DonationPayment) (domain.DonationPayment, error) {
				if payment.SenderUserID != 1002 ||
					payment.RecipientUserID != 1001 ||
					payment.AmountValue != 1500 ||
					payment.Currency != "RUB" ||
					payment.Status != domain.PaymentStatusPending ||
					payment.Provider != "stripe" ||
					payment.Message == nil ||
					*payment.Message != "Спасибо" ||
					payment.ConfirmationToken == "" {
					t.Fatalf("unexpected payment: %+v", payment)
				}
				payment.PaymentID = 81
				return payment, nil
			},
			confirmPaymentFunc: func(ctx context.Context, senderUserID int64, paymentID int64, confirmationToken string) (domain.DonationPayment, bool, error) {
				if senderUserID != 1002 || paymentID != 81 || confirmationToken != "confirm_abc" {
					t.Fatalf("unexpected confirm args: sender=%d payment=%d token=%s", senderUserID, paymentID, confirmationToken)
				}
				return domain.DonationPayment{
					PaymentID: paymentID,
					Status:    domain.PaymentStatusConfirmed,
					Donation:  &domain.Donation{DonationID: 77, SenderUserID: 1002, RecipientUserID: 1001},
				}, true, nil
			},
		}),
		nil,
		stubPaymentProvider{
			createFunc: func(ctx context.Context, request PaymentProviderCreateRequest) (PaymentProviderPayment, error) {
				if !strings.HasPrefix(request.IdempotenceKey, "content-donation-payment-confirm_") {
					t.Fatalf("unexpected idempotence key: %s", request.IdempotenceKey)
				}
				return PaymentProviderPayment{ProviderPaymentID: "provider-payment-1", Status: "pending", ConfirmationURL: "https://pay.example/1"}, nil
			},
		},
	)

	payment, err := service.CreateDonationPayment(context.Background(), CreateDonationPaymentCommand{
		SenderUserID:    1002,
		RecipientUserID: 1001,
		AmountValue:     1500,
		Message:         stringPtr(" Спасибо "),
	})
	if err != nil {
		t.Fatalf("unexpected create payment error: %v", err)
	}
	if payment.PaymentID != 81 || payment.Status != domain.PaymentStatusPending {
		t.Fatalf("unexpected payment: %+v", payment)
	}

	confirmed, err := service.ConfirmDonationPayment(context.Background(), ConfirmDonationPaymentCommand{
		SenderUserID:      1002,
		PaymentID:         81,
		ConfirmationToken: " confirm_abc ",
	})
	if err != nil {
		t.Fatalf("unexpected confirm payment error: %v", err)
	}
	if confirmed.Status != domain.PaymentStatusConfirmed || confirmed.Donation == nil || confirmed.Donation.DonationID != 77 {
		t.Fatalf("unexpected confirmed payment: %+v", confirmed)
	}
}

func TestServiceSubscriptionPaymentFlow(t *testing.T) {
	var notifications []domain.Notification
	service := NewService(
		stubRepositories(stubContentRepository{
			getTierFunc: func(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error) {
				if trainerUserID != 1001 || tierID != 2 {
					t.Fatalf("unexpected tier lookup: trainer=%d tier=%d", trainerUserID, tierID)
				}
				return domain.SubscriptionTier{TrainerUserID: trainerUserID, TierID: tierID, Name: "Продвинутый", Price: 1500}, nil
			},
			createPaymentFunc: func(ctx context.Context, payment domain.DonationPayment) (domain.DonationPayment, error) {
				if payment.SenderUserID != 1002 ||
					payment.RecipientUserID != 1001 ||
					payment.AmountValue != 1500 ||
					payment.Currency != "RUB" ||
					payment.TierID == nil ||
					*payment.TierID != 2 ||
					payment.Message != nil ||
					payment.Status != domain.PaymentStatusPending ||
					payment.Provider != "stripe" ||
					payment.ConfirmationToken == "" {
					t.Fatalf("unexpected payment: %+v", payment)
				}
				payment.PaymentID = 82
				return payment, nil
			},
			confirmPaymentFunc: func(ctx context.Context, senderUserID int64, paymentID int64, confirmationToken string) (domain.DonationPayment, bool, error) {
				if senderUserID != 1002 || paymentID != 82 || confirmationToken != "confirm_abc" {
					t.Fatalf("unexpected confirm args: sender=%d payment=%d token=%s", senderUserID, paymentID, confirmationToken)
				}
				return domain.DonationPayment{
					PaymentID: paymentID,
					Status:    domain.PaymentStatusConfirmed,
					Subscription: &domain.Subscription{
						SubscriptionID: 2401,
						ClientUserID:   1002,
						TrainerUserID:  1001,
						TierID:         2,
						Active:         true,
					},
				}, true, nil
			},
			createNotificationFunc: func(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
				notifications = append(notifications, notification)
				return notification, nil
			},
		}),
		nil,
		stubPaymentProvider{
			createFunc: func(ctx context.Context, request PaymentProviderCreateRequest) (PaymentProviderPayment, error) {
				if !strings.HasPrefix(request.IdempotenceKey, "content-subscription-payment-confirm_") {
					t.Fatalf("unexpected idempotence key: %s", request.IdempotenceKey)
				}
				return PaymentProviderPayment{ProviderPaymentID: "provider-payment-1", Status: "pending", ConfirmationURL: "https://pay.example/1"}, nil
			},
		},
	)

	payment, err := service.CreateSubscriptionPayment(context.Background(), CreateSubscriptionPaymentCommand{
		ClientUserID:  1002,
		TrainerUserID: 1001,
		TierID:        2,
	})
	if err != nil {
		t.Fatalf("unexpected create payment error: %v", err)
	}
	if payment.PaymentID != 82 || payment.Status != domain.PaymentStatusPending {
		t.Fatalf("unexpected payment: %+v", payment)
	}

	confirmed, err := service.ConfirmDonationPayment(context.Background(), ConfirmDonationPaymentCommand{
		SenderUserID:      1002,
		PaymentID:         82,
		ConfirmationToken: " confirm_abc ",
	})
	if err != nil {
		t.Fatalf("unexpected confirm payment error: %v", err)
	}
	if confirmed.Status != domain.PaymentStatusConfirmed ||
		confirmed.Subscription == nil ||
		confirmed.Subscription.SubscriptionID != 2401 {
		t.Fatalf("unexpected confirmed payment: %+v", confirmed)
	}
	if len(notifications) != 2 {
		t.Fatalf("unexpected subscription notifications: %+v", notifications)
	}
	if notifications[0].UserID != 1001 ||
		notifications[0].ActorUserID != 1002 ||
		notifications[0].Type != domain.NotificationTypeSubscription {
		t.Fatalf("unexpected trainer notification: %+v", notifications[0])
	}
	if notifications[1].UserID != 1002 ||
		notifications[1].ActorUserID != 1001 ||
		notifications[1].Type != domain.NotificationTypeSubscription {
		t.Fatalf("unexpected client notification: %+v", notifications[1])
	}
}

func TestServiceCreateSubscriptionPaymentDoesNotPersistWhenProviderFails(t *testing.T) {
	persisted := false
	service := NewService(
		stubRepositories(stubContentRepository{
			getTierFunc: func(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error) {
				return domain.SubscriptionTier{TrainerUserID: trainerUserID, TierID: tierID, Name: "Продвинутый", Price: 1500}, nil
			},
			createPaymentFunc: func(ctx context.Context, payment domain.DonationPayment) (domain.DonationPayment, error) {
				persisted = true
				return payment, nil
			},
		}),
		nil,
		stubPaymentProvider{
			createFunc: func(ctx context.Context, request PaymentProviderCreateRequest) (PaymentProviderPayment, error) {
				return PaymentProviderPayment{}, errors.New("stripe down")
			},
		},
	)

	_, err := service.CreateSubscriptionPayment(context.Background(), CreateSubscriptionPaymentCommand{
		ClientUserID:  1002,
		TrainerUserID: 1001,
		TierID:        2,
	})
	if !errors.Is(err, ErrPaymentProviderUnavailable) {
		t.Fatalf("unexpected error: %v", err)
	}
	if persisted {
		t.Fatal("expected no payment row to be created when provider fails")
	}
}

func TestServiceCreateFreeSubscriptionPaymentCancelsExistingPaidRenewal(t *testing.T) {
	periodEnd := time.Date(2026, 6, 29, 12, 0, 0, 0, time.UTC)
	canceledProviderSubscriptions := make([]string, 0)

	service := NewService(
		stubRepositories(stubContentRepository{
			getTierFunc: func(ctx context.Context, trainerUserID int64, tierID int64) (domain.SubscriptionTier, error) {
				if trainerUserID != 1001 || tierID != 1 {
					t.Fatalf("unexpected tier lookup: trainer=%d tier=%d", trainerUserID, tierID)
				}
				return domain.SubscriptionTier{
					TrainerUserID: trainerUserID,
					TierID:        tierID,
					Name:          "Бесплатный",
					Price:         0,
				}, nil
			},
			listSubscriptionsFunc: func(ctx context.Context, clientUserID int64) ([]domain.Subscription, error) {
				if clientUserID != 1002 {
					t.Fatalf("unexpected client id: %d", clientUserID)
				}
				return []domain.Subscription{{
					SubscriptionID:       2401,
					ClientUserID:         clientUserID,
					TrainerUserID:        1001,
					TierID:               2,
					TierName:             "Платный",
					Price:                1500,
					Active:               true,
					ExpiresAt:            periodEnd,
					StripeSubscriptionID: "sub_paid",
					CurrentPeriodEnd:     &periodEnd,
					AutoRenew:            true,
				}}, nil
			},
			subscribeFunc: func(ctx context.Context, subscription domain.Subscription) (domain.Subscription, error) {
				if subscription.ClientUserID != 1002 ||
					subscription.TrainerUserID != 1001 ||
					subscription.TierID != 1 ||
					subscription.TierName != "Бесплатный" ||
					subscription.Price != 0 {
					t.Fatalf("unexpected subscription: %+v", subscription)
				}
				subscription.SubscriptionID = 2401
				subscription.Active = true
				return subscription, nil
			},
		}),
		nil,
		stubPaymentProvider{
			cancelFunc: func(ctx context.Context, providerSubscriptionID string, atPeriodEnd bool) error {
				if !atPeriodEnd {
					t.Fatal("expected cancellation at period end")
				}
				canceledProviderSubscriptions = append(canceledProviderSubscriptions, providerSubscriptionID)
				return nil
			},
		},
	)

	payment, err := service.CreateSubscriptionPayment(context.Background(), CreateSubscriptionPaymentCommand{
		ClientUserID:  1002,
		TrainerUserID: 1001,
		TierID:        1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(canceledProviderSubscriptions) != 1 || canceledProviderSubscriptions[0] != "sub_paid" {
		t.Fatalf("unexpected provider cancellations: %+v", canceledProviderSubscriptions)
	}
	if payment.Status != domain.PaymentStatusConfirmed ||
		payment.Subscription == nil ||
		payment.Subscription.TierID != 1 ||
		payment.CreatedAt.IsZero() ||
		payment.UpdatedAt.IsZero() ||
		payment.ConfirmedAt == nil ||
		payment.ConfirmedAt.IsZero() {
		t.Fatalf("unexpected free subscription payment: %+v", payment)
	}
}

func TestServiceCreateDonationPaymentPersistsProviderData(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			createPaymentFunc: func(ctx context.Context, payment domain.DonationPayment) (domain.DonationPayment, error) {
				if payment.ProviderPaymentID != "provider-payment-1" ||
					payment.ConfirmationURL != "https://pay.example/1" ||
					payment.Status != domain.PaymentStatusPending {
					t.Fatalf("unexpected payment passed to repository: %+v", payment)
				}
				payment.PaymentID = 91
				return payment, nil
			},
		}),
		nil,
		stubPaymentProvider{},
	)

	payment, err := service.CreateDonationPayment(context.Background(), CreateDonationPaymentCommand{
		SenderUserID:    1002,
		RecipientUserID: 1001,
		AmountValue:     1500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payment.PaymentID != 91 || payment.ProviderPaymentID != "provider-payment-1" {
		t.Fatalf("unexpected payment: %+v", payment)
	}
}

func TestServiceSweepPayments(t *testing.T) {
	confirmed := false
	expired := false
	service := NewService(
		stubRepositories(stubContentRepository{
			listStalePaymentsFunc: func(ctx context.Context, olderThan time.Time, limit int32) ([]domain.DonationPayment, error) {
				return []domain.DonationPayment{
					{PaymentID: 1, ProviderPaymentID: "cs_paid", Status: domain.PaymentStatusPending},
					{PaymentID: 2, ProviderPaymentID: "cs_dead", Status: domain.PaymentStatusPending},
				}, nil
			},
			confirmByProviderFunc: func(ctx context.Context, providerPaymentID string) (domain.DonationPayment, bool, error) {
				if providerPaymentID != "cs_paid" {
					t.Fatalf("unexpected confirm provider id: %s", providerPaymentID)
				}
				confirmed = true
				return domain.DonationPayment{ProviderPaymentID: providerPaymentID, Status: domain.PaymentStatusConfirmed}, true, nil
			},
			setPaymentStatusFunc: func(ctx context.Context, providerPaymentID string, status domain.PaymentStatus) (domain.DonationPayment, bool, error) {
				if providerPaymentID != "cs_dead" || status != domain.PaymentStatusExpired {
					t.Fatalf("unexpected set status: id=%s status=%s", providerPaymentID, status)
				}
				expired = true
				return domain.DonationPayment{ProviderPaymentID: providerPaymentID, Status: status}, true, nil
			},
			deactivateExpiredFunc: func(ctx context.Context, expiredBefore time.Time) (int64, error) {
				return 3, nil
			},
		}),
		nil,
		stubPaymentProvider{
			getFunc: func(ctx context.Context, providerPaymentID string) (PaymentProviderPayment, error) {
				if providerPaymentID == "cs_paid" {
					return PaymentProviderPayment{ProviderPaymentID: providerPaymentID, Status: "succeeded"}, nil
				}
				return PaymentProviderPayment{ProviderPaymentID: providerPaymentID, Status: "canceled"}, nil
			},
		},
	)

	result, err := service.SweepPayments(context.Background(), 24*time.Hour, 72*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PendingChecked != 2 || result.Confirmed != 1 || result.Expired != 1 || result.Errors != 0 {
		t.Fatalf("unexpected sweep result: %+v", result)
	}
	if result.SubscriptionsDeactivated != 3 {
		t.Fatalf("unexpected deactivated count: %d", result.SubscriptionsDeactivated)
	}
	if !confirmed || !expired {
		t.Fatalf("expected confirm and expire to be called: confirmed=%v expired=%v", confirmed, expired)
	}
}

func TestServiceExpireAndFailPaymentFromProvider(t *testing.T) {
	var recorded []domain.PaymentStatus
	service := NewService(
		stubRepositories(stubContentRepository{
			setPaymentStatusFunc: func(ctx context.Context, providerPaymentID string, status domain.PaymentStatus) (domain.DonationPayment, bool, error) {
				if providerPaymentID != "cs_test_1" {
					t.Fatalf("unexpected provider payment id: %s", providerPaymentID)
				}
				recorded = append(recorded, status)
				return domain.DonationPayment{ProviderPaymentID: providerPaymentID, Status: status}, true, nil
			},
		}),
		nil,
	)

	expired, err := service.ExpirePaymentFromProvider(context.Background(), "cs_test_1")
	if err != nil {
		t.Fatalf("unexpected expire error: %v", err)
	}
	if expired.Status != domain.PaymentStatusExpired {
		t.Fatalf("unexpected expired status: %s", expired.Status)
	}

	failed, err := service.FailPaymentFromProvider(context.Background(), "cs_test_1")
	if err != nil {
		t.Fatalf("unexpected fail error: %v", err)
	}
	if failed.Status != domain.PaymentStatusFailed {
		t.Fatalf("unexpected failed status: %s", failed.Status)
	}

	if len(recorded) != 2 || recorded[0] != domain.PaymentStatusExpired || recorded[1] != domain.PaymentStatusFailed {
		t.Fatalf("unexpected recorded statuses: %v", recorded)
	}
}

func TestServiceSetTerminalPaymentStatusRejectsEmptyProviderID(t *testing.T) {
	service := NewService(stubRepositories(stubContentRepository{}), nil)

	if _, err := service.FailPaymentFromProvider(context.Background(), "  "); !errors.Is(err, ErrInvalidProviderPaymentID) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceGetBalance(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			getBalanceFunc: func(ctx context.Context, trainerUserID int64, currency string) (domain.Balance, error) {
				if trainerUserID != 1001 || currency != "RUB" {
					t.Fatalf("unexpected balance args: trainer=%d currency=%s", trainerUserID, currency)
				}
				return domain.Balance{TrainerUserID: trainerUserID, AmountValue: 3000, Currency: currency}, nil
			},
		}),
		nil,
	)

	balance, err := service.GetBalance(context.Background(), GetBalanceQuery{TrainerUserID: 1001})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance.AmountValue != 3000 || balance.Currency != "RUB" {
		t.Fatalf("unexpected balance: %+v", balance)
	}
}

func TestServiceGetTrainerStatistics(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			getStatisticsFunc: func(ctx context.Context, trainerUserID int64, currency string, monthStart time.Time) (domain.TrainerStatistics, error) {
				if trainerUserID != 1001 || currency != "RUB" {
					t.Fatalf("unexpected statistics args: trainer=%d currency=%s", trainerUserID, currency)
				}
				if monthStart.Day() != 1 || monthStart.Hour() != 0 || monthStart.Minute() != 0 {
					t.Fatalf("unexpected month start: %s", monthStart)
				}
				return domain.TrainerStatistics{
					TrainerUserID:  trainerUserID,
					PostsCount:     12,
					DonationsCount: 4,
					TotalRevenue:   7000,
					MonthlyRevenue: 2500,
					Currency:       currency,
				}, nil
			},
		}),
		nil,
	)

	statistics, err := service.GetTrainerStatistics(context.Background(), GetTrainerStatisticsQuery{TrainerUserID: 1001})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if statistics.PostsCount != 12 || statistics.MonthlyRevenue != 2500 || statistics.Currency != "RUB" {
		t.Fatalf("unexpected statistics: %+v", statistics)
	}
}

func TestServiceDeletePost(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{PostID: postID, AuthorUserID: viewerUserID}, nil
			},
			deletePostFunc: func(ctx context.Context, postID int64, authorUserID int64) error {
				if postID != 33 || authorUserID != 7 {
					t.Fatalf("unexpected delete args: post=%d author=%d", postID, authorUserID)
				}
				return nil
			},
		}),
		nil,
	)

	if err := service.DeletePost(context.Background(), DeletePostCommand{PostID: 33, AuthorUserID: 7}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestServiceLikeAndUnlikePost(t *testing.T) {
	requiredLevel := int32(1)
	var notifications []domain.Notification
	service := NewService(
		stubRepositories(stubContentRepository{
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{PostID: postID, AuthorUserID: 7, RequiredSubscriptionLevel: &requiredLevel}, nil
			},
			activeLevelFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error) {
				return &requiredLevel, nil
			},
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) (bool, error) {
				if postID != 33 || userID != 13 {
					t.Fatalf("unexpected like args: post=%d user=%d", postID, userID)
				}
				return true, nil
			},
			deleteLikeFunc: func(ctx context.Context, postID int64, userID int64) error {
				if postID != 33 || userID != 13 {
					t.Fatalf("unexpected unlike args: post=%d user=%d", postID, userID)
				}
				return nil
			},
			getLikeStateFunc: func(ctx context.Context, postID int64, userID int64) (domain.PostLikeState, error) {
				return domain.PostLikeState{PostID: postID, LikesCount: 1, IsLiked: true}, nil
			},
			createNotificationFunc: func(ctx context.Context, notification domain.Notification) (domain.Notification, error) {
				notifications = append(notifications, notification)
				return notification, nil
			},
		}),
		nil,
	)

	state, err := service.LikePost(context.Background(), LikePostCommand{PostID: 33, UserID: 13})
	if err != nil {
		t.Fatalf("unexpected like error: %v", err)
	}
	if state.PostID != 33 || state.LikesCount != 1 {
		t.Fatalf("unexpected like state: %+v", state)
	}
	if len(notifications) != 1 ||
		notifications[0].Type != domain.NotificationTypeLike ||
		notifications[0].UserID != 7 ||
		notifications[0].ActorUserID != 13 ||
		notifications[0].PostID == nil ||
		*notifications[0].PostID != 33 {
		t.Fatalf("unexpected like notifications: %+v", notifications)
	}
	state, err = service.UnlikePost(context.Background(), LikePostCommand{PostID: 33, UserID: 13})
	if err != nil {
		t.Fatalf("unexpected unlike error: %v", err)
	}
	if state.PostID != 33 {
		t.Fatalf("unexpected unlike state: %+v", state)
	}
}

func TestServiceGetPostRejectsRestrictedAccess(t *testing.T) {
	requiredLevel := int32(2)

	service := NewService(
		stubRepositories(stubContentRepository{
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
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) (bool, error) { return true, nil },
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
		}),
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
		stubRepositories(stubContentRepository{
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
			upsertLikeFunc: func(ctx context.Context, postID int64, userID int64) (bool, error) { return true, nil },
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
		}),
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

func TestServiceListComments(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			getPostFunc: func(ctx context.Context, postID int64, viewerUserID int64) (domain.Post, error) {
				return domain.Post{PostID: postID, AuthorUserID: 7}, nil
			},
			listCommentsFunc: func(ctx context.Context, postID int64, limit int32, offset int32) ([]domain.Comment, error) {
				if postID != 21 || limit != 20 || offset != 5 {
					t.Fatalf("unexpected list comments args: post=%d limit=%d offset=%d", postID, limit, offset)
				}
				return []domain.Comment{{CommentID: 1, PostID: postID, Body: "Great"}}, nil
			},
		}),
		nil,
	)

	comments, err := service.ListComments(context.Background(), ListCommentsQuery{
		PostID:       21,
		ViewerUserID: 13,
		Limit:        0,
		Offset:       5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(comments) != 1 {
		t.Fatalf("unexpected comments: %+v", comments)
	}
}

func stringPtr(value string) *string {
	return &value
}

func TestServiceListAuthorPosts(t *testing.T) {
	requiredLevel := int32(1)
	service := NewService(
		stubRepositories(stubContentRepository{
			listAuthorPostsFunc: func(ctx context.Context, authorUserID int64, viewerUserID int64, limit int32, offset int32) ([]domain.PostSummary, error) {
				if authorUserID != 7 {
					t.Fatalf("unexpected author user id: %d", authorUserID)
				}
				return []domain.PostSummary{
					{PostID: 101, AuthorUserID: 7, Title: "Run Day 1", RequiredSubscriptionLevel: &requiredLevel, CanView: false},
					{PostID: 102, AuthorUserID: 7, Title: "Run Day 2", RequiredSubscriptionLevel: nil, CanView: true},
				}, nil
			},
			activeLevelFunc: func(ctx context.Context, clientUserID int64, trainerUserID int64) (*int32, error) {
				level := int32(2)
				return &level, nil
			},
		}),
		nil,
	)

	posts, err := service.ListAuthorPosts(context.Background(), ListAuthorPostsQuery{
		AuthorUserID: 7,
		ViewerUserID: 3,
		Limit:        10,
		Offset:       0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("unexpected posts count: %d", len(posts))
	}
}

func TestServiceListAuthorPostsInvalidID(t *testing.T) {
	service := NewService(stubRepositories(stubContentRepository{}), nil)

	_, err := service.ListAuthorPosts(context.Background(), ListAuthorPostsQuery{
		AuthorUserID: 0,
		ViewerUserID: 3,
	})
	if err == nil {
		t.Fatal("expected error for invalid author id")
	}
}

func TestServiceListNotifications(t *testing.T) {
	service := NewService(
		stubRepositories(stubContentRepository{
			listNotificationsFunc: func(ctx context.Context, userID int64, limit int32, offset int32) ([]domain.Notification, error) {
				if userID != 5 {
					t.Fatalf("unexpected user id: %d", userID)
				}
				return []domain.Notification{
					{NotificationID: 1, UserID: 5, Type: "new_post"},
					{NotificationID: 2, UserID: 5, Type: "donation"},
				}, nil
			},
		}),
		nil,
	)

	notifications, err := service.ListNotifications(context.Background(), ListNotificationsQuery{
		UserID: 5,
		Limit:  20,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifications) != 2 {
		t.Fatalf("unexpected notifications count: %d", len(notifications))
	}
}

func TestServiceListNotificationsInvalidUserID(t *testing.T) {
	service := NewService(stubRepositories(stubContentRepository{}), nil)

	_, err := service.ListNotifications(context.Background(), ListNotificationsQuery{UserID: 0})
	if err == nil {
		t.Fatal("expected error for zero user id")
	}
}

func TestServiceMarkNotificationRead(t *testing.T) {
	markCalled := false
	service := NewService(
		stubRepositories(stubContentRepository{
			markNotificationFunc: func(ctx context.Context, userID int64, notificationID int64) (domain.Notification, error) {
				markCalled = true
				return domain.Notification{NotificationID: notificationID, UserID: userID, Type: "new_post"}, nil
			},
		}),
		nil,
	)

	notification, err := service.MarkNotificationRead(context.Background(), MarkNotificationReadCommand{
		UserID:         5,
		NotificationID: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !markCalled {
		t.Fatal("expected repository mark to be called")
	}
	if notification.NotificationID != 10 {
		t.Fatalf("unexpected notification id: %d", notification.NotificationID)
	}
}

func TestServiceMarkNotificationReadInvalidIDs(t *testing.T) {
	service := NewService(stubRepositories(stubContentRepository{}), nil)

	_, err := service.MarkNotificationRead(context.Background(), MarkNotificationReadCommand{UserID: 0, NotificationID: 5})
	if err == nil {
		t.Fatal("expected error for zero user id")
	}

	_, err = service.MarkNotificationRead(context.Background(), MarkNotificationReadCommand{UserID: 5, NotificationID: 0})
	if err == nil {
		t.Fatal("expected error for zero notification id")
	}
}
