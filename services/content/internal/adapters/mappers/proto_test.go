package mappers

import (
	"errors"
	"testing"
	"time"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestContentRequestMappers(t *testing.T) {
	level := int32(2)
	sportTypeID := int64(3001)
	text := "text"
	fileURL := "http://cdn/image.jpg"
	description := "desc"
	message := "message"

	create := CreatePostRequestToCommand(&contentv1.CreatePostRequest{
		AuthorUserId:              1001,
		Title:                     "Post",
		RequiredSubscriptionLevel: &level,
		SportTypeId:               &sportTypeID,
		Blocks: []*contentv1.PostBlockInput{
			{Kind: contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT, TextContent: &text},
			{Kind: contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE, FileUrl: &fileURL},
		},
	})
	if create.AuthorUserID != 1001 ||
		create.RequiredSubscriptionLevel == nil ||
		*create.RequiredSubscriptionLevel != 2 ||
		create.SportTypeID == nil ||
		*create.SportTypeID != 3001 ||
		len(create.Blocks) != 2 ||
		create.Blocks[0].Kind != domain.BlockKindText ||
		create.Blocks[1].Kind != domain.BlockKindImage {
		t.Fatalf("unexpected create command: %+v", create)
	}

	search := SearchPostsRequestToQuery(&contentv1.SearchPostsRequest{
		Query:                   "run",
		AuthorUserIds:           []int64{1001},
		SportTypeIds:            []int64{3001},
		BlockKinds:              []contentv1.ContentBlockKind{contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT},
		OnlyAvailable:           true,
		ViewerUserId:            1002,
		ViewerSubscriptionLevel: &level,
		Limit:                   20,
		Offset:                  5,
	})
	if search.Query != "run" ||
		len(search.AuthorUserIDs) != 1 ||
		search.BlockKinds[0] != domain.BlockKindDocument ||
		search.ViewerSubscriptionLevel == nil ||
		*search.ViewerSubscriptionLevel != 2 {
		t.Fatalf("unexpected search query: %+v", search)
	}

	update := UpdatePostRequestToCommand(&contentv1.UpdatePostRequest{
		PostId:                         101,
		AuthorUserId:                   1001,
		Title:                          stringPtr("Updated"),
		RequiredSubscriptionLevel:      &level,
		ClearRequiredSubscriptionLevel: true,
		SportTypeId:                    &sportTypeID,
		ClearSportTypeId:               true,
		ReplaceBlocks:                  true,
		Blocks:                         []*contentv1.PostBlockInput{{Kind: contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT, TextContent: &text}},
	})
	if update.PostID != 101 || update.Title == nil || !update.ClearSportTypeID || !update.ReplaceBlocks {
		t.Fatalf("unexpected update command: %+v", update)
	}

	tier := CreateSubscriptionTierRequestToCommand(&contentv1.CreateSubscriptionTierRequest{
		TrainerUserId: 1001,
		Name:          "Basic",
		Price:         500,
		Description:   &description,
	})
	if tier.TrainerUserID != 1001 || tier.Description == nil || *tier.Description != "desc" {
		t.Fatalf("unexpected tier command: %+v", tier)
	}

	donation := DonateToProfileRequestToCommand(&contentv1.DonateToProfileRequest{
		SenderUserId:    1002,
		RecipientUserId: 1001,
		AmountValue:     1500,
		Currency:        "RUB",
		Message:         &message,
	})
	if donation.SenderUserID != 1002 || donation.RecipientUserID != 1001 || donation.Message == nil {
		t.Fatalf("unexpected donation command: %+v", donation)
	}
}

func TestContentOtherRequestMappers(t *testing.T) {
	level := int32(1)

	if got := ListAuthorPostsRequestToQuery(&contentv1.ListAuthorPostsRequest{AuthorUserId: 1, ViewerUserId: 2, ViewerSubscriptionLevel: &level, Limit: 10, Offset: 5}); got.AuthorUserID != 1 || got.ViewerUserID != 2 {
		t.Fatalf("unexpected list author query: %+v", got)
	}
	if got := UploadPostMediaRequestToCommand(&contentv1.UploadPostMediaRequest{AuthorUserId: 1, FileName: "a.jpg", ContentType: "image/jpeg", File: []byte("x")}); got.AuthorUserID != 1 || got.FileName != "a.jpg" {
		t.Fatalf("unexpected upload command: %+v", got)
	}
	if got := GetPostRequestToQuery(&contentv1.GetPostRequest{PostId: 10, ViewerUserId: 2, ViewerSubscriptionLevel: &level}); got.PostID != 10 || got.ViewerSubscriptionLevel == nil {
		t.Fatalf("unexpected get post query: %+v", got)
	}
	if got := ListSubscriptionTiersRequestToQuery(&contentv1.ListSubscriptionTiersRequest{TrainerUserId: 1}); got.TrainerUserID != 1 {
		t.Fatalf("unexpected list tiers query: %+v", got)
	}
	if got := UpdateSubscriptionTierRequestToCommand(&contentv1.UpdateSubscriptionTierRequest{TrainerUserId: 1, TierId: 2, Name: stringPtr("n"), Price: int32Ptr(3), Description: stringPtr("d"), ClearDescription: true}); got.TierID != 2 || got.Price == nil || !got.ClearDescription {
		t.Fatalf("unexpected update tier command: %+v", got)
	}
	if got := DeleteSubscriptionTierRequestToCommand(&contentv1.DeleteSubscriptionTierRequest{TrainerUserId: 1, TierId: 2}); got.TierID != 2 {
		t.Fatalf("unexpected delete tier command: %+v", got)
	}
	if got := SubscribeToTrainerRequestToCommand(&contentv1.SubscribeToTrainerRequest{ClientUserId: 1, TrainerUserId: 2, TierId: 3}); got.TierID != 3 {
		t.Fatalf("unexpected subscribe command: %+v", got)
	}
	if got := ListMySubscriptionsRequestToQuery(&contentv1.ListMySubscriptionsRequest{ClientUserId: 1}); got.ClientUserID != 1 {
		t.Fatalf("unexpected list subscriptions query: %+v", got)
	}
	if got := UpdateSubscriptionRequestToCommand(&contentv1.UpdateSubscriptionRequest{ClientUserId: 1, SubscriptionId: 2, TierId: 3}); got.SubscriptionID != 2 {
		t.Fatalf("unexpected update subscription command: %+v", got)
	}
	if got := CancelSubscriptionRequestToCommand(&contentv1.CancelSubscriptionRequest{ClientUserId: 1, SubscriptionId: 2}); got.SubscriptionID != 2 {
		t.Fatalf("unexpected cancel subscription command: %+v", got)
	}
	if got := GetBalanceRequestToQuery(&contentv1.GetBalanceRequest{TrainerUserId: 1, Currency: "RUB"}); got.TrainerUserID != 1 || got.Currency != "RUB" {
		t.Fatalf("unexpected balance query: %+v", got)
	}
	if got := DeletePostRequestToCommand(&contentv1.DeletePostRequest{PostId: 1, AuthorUserId: 2}); got.PostID != 1 || got.AuthorUserID != 2 {
		t.Fatalf("unexpected delete post command: %+v", got)
	}
	if got := LikePostRequestToCommand(&contentv1.LikePostRequest{PostId: 1, UserId: 2, ViewerSubscriptionLevel: &level}); got.PostID != 1 || got.ViewerSubscriptionLevel == nil {
		t.Fatalf("unexpected like command: %+v", got)
	}
	if got := UnlikePostRequestToCommand(&contentv1.UnlikePostRequest{PostId: 1, UserId: 2, ViewerSubscriptionLevel: &level}); got.PostID != 1 || got.ViewerSubscriptionLevel == nil {
		t.Fatalf("unexpected unlike command: %+v", got)
	}
	if got := CreateCommentRequestToCommand(&contentv1.CreateCommentRequest{PostId: 1, AuthorUserId: 2, ViewerSubscriptionLevel: &level, Body: "body"}); got.Body != "body" {
		t.Fatalf("unexpected create comment command: %+v", got)
	}
	if got := ListCommentsRequestToQuery(&contentv1.ListCommentsRequest{PostId: 1, ViewerUserId: 2, ViewerSubscriptionLevel: &level, Limit: 10, Offset: 5}); got.PostID != 1 || got.Limit != 10 {
		t.Fatalf("unexpected list comments query: %+v", got)
	}
}

func TestContentResponseMappers(t *testing.T) {
	now := time.Date(2026, time.May, 6, 12, 0, 0, 0, time.UTC)
	level := int32(2)
	sportTypeID := int64(3001)
	text := "text"
	fileURL := "http://cdn/image.jpg"

	post := domain.Post{
		PostID:                    101,
		AuthorUserID:              1001,
		Title:                     "Post",
		RequiredSubscriptionLevel: &level,
		SportTypeID:               &sportTypeID,
		CanView:                   true,
		LikesCount:                5,
		IsLiked:                   true,
		CommentsCount:             2,
		CreatedAt:                 now,
		UpdatedAt:                 now,
		Blocks: []domain.PostBlock{
			{PostBlockID: 1, Position: 0, Kind: domain.BlockKindText, TextContent: &text},
			{PostBlockID: 2, Position: 1, Kind: domain.BlockKindImage, FileURL: &fileURL},
		},
	}
	postResponse := NewPostResponse(post)
	if postResponse.GetPost().GetPostId() != 101 ||
		postResponse.GetPost().GetRequiredSubscriptionLevel() != 2 ||
		len(postResponse.GetPost().GetBlocks()) != 2 ||
		postResponse.GetPost().GetBlocks()[1].GetKind() != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE {
		t.Fatalf("unexpected post response: %+v", postResponse)
	}

	posts := NewSearchPostsResponse([]domain.PostSummary{{
		PostID:                    101,
		AuthorUserID:              1001,
		Title:                     "Post",
		RequiredSubscriptionLevel: &level,
		SportTypeID:               &sportTypeID,
		CanView:                   true,
		CreatedAt:                 now,
	}})
	if len(posts.GetPosts()) != 1 || posts.GetPosts()[0].GetSportTypeId() != 3001 {
		t.Fatalf("unexpected posts response: %+v", posts)
	}
	if len(NewListAuthorPostsResponse([]domain.PostSummary{{PostID: 1}}).GetPosts()) != 1 {
		t.Fatal("unexpected author posts response")
	}
}

func TestContentMoreResponseMappers(t *testing.T) {
	now := time.Date(2026, time.May, 6, 12, 0, 0, 0, time.UTC)
	description := "desc"
	message := "message"

	if got := NewPostMediaResponse(domain.PostMedia{FileURL: "url", Kind: domain.BlockKindVideo, ContentType: "video/mp4", SizeBytes: 10}); got.GetMedia().GetKind() != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO {
		t.Fatalf("unexpected media response: %+v", got)
	}
	if got := NewPostLikeStateResponse(domain.PostLikeState{PostID: 1, LikesCount: 2, IsLiked: true}); got.GetState().GetLikesCount() != 2 {
		t.Fatalf("unexpected like state response: %+v", got)
	}
	if got := NewListSubscriptionTiersResponse([]domain.SubscriptionTier{{TierID: 1, TrainerUserID: 2, Name: "Basic", Price: 500, Description: &description, CreatedAt: now, UpdatedAt: now}}); len(got.GetTiers()) != 1 || got.GetTiers()[0].GetDescription() != "desc" {
		t.Fatalf("unexpected tiers response: %+v", got)
	}
	if got := NewSubscriptionTierResponse(domain.SubscriptionTier{TierID: 1, TrainerUserID: 2, Name: "Basic", Price: 500}); got.GetTierId() != 1 {
		t.Fatalf("unexpected tier response: %+v", got)
	}
	subscription := domain.Subscription{SubscriptionID: 1, ClientUserID: 2, TrainerUserID: 3, TierID: 4, TierName: "Basic", Price: 500, Active: true, ExpiresAt: now, CreatedAt: now, UpdatedAt: now}
	if got := NewSubscriptionResponse(subscription); got.GetSubscriptionId() != 1 || !got.GetActive() {
		t.Fatalf("unexpected subscription response: %+v", got)
	}
	if got := NewListMySubscriptionsResponse([]domain.Subscription{subscription}); len(got.GetSubscriptions()) != 1 {
		t.Fatalf("unexpected subscriptions response: %+v", got)
	}
	if got := NewDonationResponse(domain.Donation{DonationID: 1, SenderUserID: 2, RecipientUserID: 3, AmountValue: 100, Currency: "RUB", Message: &message, CreatedAt: now}); got.GetDonation().GetMessage() != "message" {
		t.Fatalf("unexpected donation response: %+v", got)
	}
	if got := NewBalanceResponse(domain.Balance{TrainerUserID: 1, AmountValue: 100, Currency: "RUB"}); got.GetAmountValue() != 100 {
		t.Fatalf("unexpected balance response: %+v", got)
	}
	if got := NewCommentResponse(domain.Comment{CommentID: 1, PostID: 2, AuthorUserID: 3, Body: "body", CreatedAt: now, UpdatedAt: now}); got.GetComment().GetBody() != "body" {
		t.Fatalf("unexpected comment response: %+v", got)
	}
	if got := NewListCommentsResponse([]domain.Comment{{CommentID: 1, Body: "body"}}); len(got.GetComments()) != 1 {
		t.Fatalf("unexpected comments response: %+v", got)
	}
	if Empty() == nil {
		t.Fatal("empty response is nil")
	}
}

func TestContentErrorToStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code codes.Code
	}{
		{name: "nil", err: nil, code: codes.OK},
		{name: "invalid argument", err: usecase.ErrInvalidPostID, code: codes.InvalidArgument},
		{name: "not found", err: domain.ErrPostNotFound, code: codes.NotFound},
		{name: "permission denied", err: domain.ErrPostForbidden, code: codes.PermissionDenied},
		{name: "failed precondition", err: domain.ErrSubscriptionTierInUse, code: codes.FailedPrecondition},
		{name: "unavailable", err: usecase.ErrPostMediaStorageUnavailable, code: codes.Unavailable},
		{name: "internal", err: errors.New("boom"), code: codes.Internal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ErrorToStatus(test.err)
			if test.err == nil {
				if err != nil {
					t.Fatalf("expected nil error, got %v", err)
				}
				return
			}
			if got := status.Code(err); got != test.code {
				t.Fatalf("code = %s, want %s", got, test.code)
			}
		})
	}
}

func TestContentBlockKindFallbacks(t *testing.T) {
	request := CreatePostRequestToCommand(&contentv1.CreatePostRequest{
		Blocks: []*contentv1.PostBlockInput{{Kind: contentv1.ContentBlockKind(99)}},
	})
	if len(request.Blocks) != 1 || request.Blocks[0].Kind != "" {
		t.Fatalf("unexpected unknown block kind: %+v", request.Blocks)
	}

	response := NewPostMediaResponse(domain.PostMedia{Kind: domain.BlockKind("unknown")})
	if response.GetMedia().GetKind() != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_UNSPECIFIED {
		t.Fatalf("unexpected unknown proto kind: %+v", response)
	}
}

func stringPtr(value string) *string {
	return &value
}

func int32Ptr(value int32) *int32 {
	return &value
}
