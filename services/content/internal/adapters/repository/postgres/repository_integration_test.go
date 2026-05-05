package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	_ "github.com/lib/pq"
)

const repositorySchemaSQL = `
DROP TABLE IF EXISTS content_post_like;
DROP TABLE IF EXISTS content_comment;
DROP TABLE IF EXISTS content_post_block;
DROP TABLE IF EXISTS content_post;
DROP TABLE IF EXISTS content_subscription;
DROP TABLE IF EXISTS content_subscription_tier;
DROP TYPE IF EXISTS content_block_kind;

CREATE TYPE content_block_kind AS ENUM ('text', 'image', 'video', 'document');

CREATE TABLE content_subscription_tier (
	trainer_user_id BIGINT NOT NULL,
	tier_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	price INTEGER NOT NULL,
	description TEXT,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	PRIMARY KEY (trainer_user_id, tier_id)
);

CREATE TABLE content_post (
	post_id BIGSERIAL PRIMARY KEY,
	author_user_id BIGINT NOT NULL,
	title TEXT NOT NULL,
	required_subscription_level INTEGER,
	sport_type_id BIGINT,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	FOREIGN KEY (author_user_id, required_subscription_level)
		REFERENCES content_subscription_tier(trainer_user_id, tier_id)
);

CREATE TABLE content_post_block (
	post_block_id BIGSERIAL PRIMARY KEY,
	post_id BIGINT NOT NULL REFERENCES content_post(post_id) ON DELETE CASCADE,
	position INTEGER NOT NULL,
	kind content_block_kind NOT NULL,
	text_content TEXT,
	file_url TEXT,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	UNIQUE (post_id, position)
);

CREATE TABLE content_comment (
	comment_id BIGSERIAL PRIMARY KEY,
	post_id BIGINT NOT NULL REFERENCES content_post(post_id) ON DELETE CASCADE,
	author_user_id BIGINT NOT NULL,
	body TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE content_post_like (
	post_id BIGINT NOT NULL REFERENCES content_post(post_id) ON DELETE CASCADE,
	user_id BIGINT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	PRIMARY KEY (post_id, user_id)
);

CREATE TABLE content_subscription (
	subscription_id BIGSERIAL PRIMARY KEY,
	client_user_id BIGINT NOT NULL,
	trainer_user_id BIGINT NOT NULL,
	tier_id INTEGER NOT NULL,
	active BOOLEAN NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	updated_at TIMESTAMPTZ NOT NULL,
	FOREIGN KEY (trainer_user_id, tier_id)
		REFERENCES content_subscription_tier(trainer_user_id, tier_id)
);

INSERT INTO content_subscription_tier (trainer_user_id, tier_id, name, price, description, created_at, updated_at)
VALUES
	(7, 1, 'VIP', 3000, 'Самый дорогой уровень', now(), now()),
	(7, 2, 'Продвинутый', 1500, 'Закрытые тренировки', now(), now());
`

func TestRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("CONTENT_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("CONTENT_TEST_DATABASE_DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(repositorySchemaSQL); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	repository := postgres.NewRepository(db)
	requiredLevel := int32(2)
	sportTypeID := int64(3001)

	postID, err := repository.CreatePost(context.Background(), domain.Post{
		AuthorUserID:              7,
		Title:                     "Morning run",
		RequiredSubscriptionLevel: &requiredLevel,
		SportTypeID:               &sportTypeID,
		Blocks: []domain.PostBlock{
			{
				Position:    0,
				Kind:        domain.BlockKindText,
				TextContent: stringPtr("Warm-up block"),
			},
			{
				Position: 1,
				Kind:     domain.BlockKindImage,
				FileURL:  stringPtr("https://cdn.example/run.jpg"),
			},
		},
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	post, err := repository.GetPost(context.Background(), postID, 55)
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	if post.Title != "Morning run" || post.SportTypeID == nil || *post.SportTypeID != 3001 || len(post.Blocks) != 2 {
		t.Fatalf("unexpected post: %+v", post)
	}

	if err := repository.UpsertLike(context.Background(), postID, 55); err != nil {
		t.Fatalf("upsert like: %v", err)
	}

	state, err := repository.GetPostLikeState(context.Background(), postID, 55)
	if err != nil {
		t.Fatalf("get like state: %v", err)
	}
	if state.LikesCount != 1 || !state.IsLiked {
		t.Fatalf("unexpected like state: %+v", state)
	}

	comment, err := repository.CreateComment(context.Background(), domain.Comment{
		PostID:       postID,
		AuthorUserID: 55,
		Body:         "Great workout",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if comment.CommentID == 0 {
		t.Fatalf("unexpected comment: %+v", comment)
	}

	comments, err := repository.ListComments(context.Background(), postID, 10, 0)
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(comments) != 1 || comments[0].Body != "Great workout" {
		t.Fatalf("unexpected comments: %+v", comments)
	}

	post.Title = "Evening run"
	post.RequiredSubscriptionLevel = nil
	post.Blocks = []domain.PostBlock{{
		Position:    0,
		Kind:        domain.BlockKindText,
		TextContent: stringPtr("Updated block"),
	}}
	if err := repository.UpdatePost(context.Background(), post, true); err != nil {
		t.Fatalf("update post: %v", err)
	}

	updatedPost, err := repository.GetPost(context.Background(), postID, 55)
	if err != nil {
		t.Fatalf("get updated post: %v", err)
	}
	if updatedPost.Title != "Evening run" || len(updatedPost.Blocks) != 1 || updatedPost.CommentsCount != 1 || !updatedPost.IsLiked {
		t.Fatalf("unexpected updated post: %+v", updatedPost)
	}

	posts, err := repository.ListAuthorPosts(context.Background(), 7, 55, 10, 0)
	if err != nil {
		t.Fatalf("list author posts: %v", err)
	}
	if len(posts) != 1 || posts[0].PostID != postID {
		t.Fatalf("unexpected author posts: %+v", posts)
	}

	searchPosts, err := repository.SearchPosts(context.Background(), usecase.SearchPostsQuery{
		Query:         "updated",
		AuthorUserIDs: []int64{7},
		SportTypeIDs:  []int64{3001},
		BlockKinds:    []domain.BlockKind{domain.BlockKindText},
		OnlyAvailable: true,
		ViewerUserID:  55,
		Limit:         10,
	})
	if err != nil {
		t.Fatalf("search posts: %v", err)
	}
	if len(searchPosts) != 1 || searchPosts[0].PostID != postID || !searchPosts[0].IsLiked || searchPosts[0].CommentsCount != 1 {
		t.Fatalf("unexpected search posts: %+v", searchPosts)
	}

	tiers, err := repository.ListSubscriptionTiers(context.Background(), 7)
	if err != nil {
		t.Fatalf("list tiers: %v", err)
	}
	if len(tiers) != 2 || tiers[0].TierID != 2 || tiers[1].TierID != 1 {
		t.Fatalf("unexpected tiers: %+v", tiers)
	}

	createdTier, err := repository.CreateSubscriptionTier(context.Background(), domain.SubscriptionTier{
		TrainerUserID: 7,
		Name:          "Премиум",
		Price:         2500,
		Description:   stringPtr("Персональные разборы"),
	})
	if err != nil {
		t.Fatalf("create tier: %v", err)
	}
	if createdTier.TierID != 3 {
		t.Fatalf("unexpected created tier: %+v", createdTier)
	}

	subscription, err := repository.SubscribeToTrainer(context.Background(), domain.Subscription{
		ClientUserID:  55,
		TrainerUserID: 7,
		TierID:        2,
		ExpiresAt:     updatedPost.CreatedAt.AddDate(0, 1, 0),
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	if subscription.SubscriptionID == 0 || !subscription.Active || subscription.TierName != "Продвинутый" {
		t.Fatalf("unexpected subscription: %+v", subscription)
	}

	subscription, err = repository.UpdateSubscription(context.Background(), domain.Subscription{
		SubscriptionID: subscription.SubscriptionID,
		ClientUserID:   55,
		TierID:         createdTier.TierID,
	})
	if err != nil {
		t.Fatalf("update subscription: %v", err)
	}
	if subscription.TierID != createdTier.TierID || subscription.TierName != "Премиум" || subscription.Price != 2500 {
		t.Fatalf("unexpected updated subscription: %+v", subscription)
	}

	level, err := repository.GetActiveSubscriptionLevel(context.Background(), 55, 7)
	if err != nil {
		t.Fatalf("active subscription level: %v", err)
	}
	if level == nil || *level != 3 {
		t.Fatalf("unexpected active subscription level: %v", level)
	}

	subscriptions, err := repository.ListSubscriptions(context.Background(), 55)
	if err != nil {
		t.Fatalf("list subscriptions: %v", err)
	}
	if len(subscriptions) != 1 || subscriptions[0].SubscriptionID != subscription.SubscriptionID {
		t.Fatalf("unexpected subscriptions: %+v", subscriptions)
	}

	if err := repository.CancelSubscription(context.Background(), 55, subscription.SubscriptionID); err != nil {
		t.Fatalf("cancel subscription: %v", err)
	}

	if err := repository.DeletePost(context.Background(), postID, 7); err != nil {
		t.Fatalf("delete post: %v", err)
	}

	if _, err := repository.GetPost(context.Background(), postID, 55); !errors.Is(err, domain.ErrPostNotFound) {
		t.Fatalf("unexpected error after delete: %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}
