package mappers

import (
	"testing"
	"time"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreatePostRequestToContentPreservesBlockOrder(t *testing.T) {
	sportTypeID := int32(3001)
	request := &gatewayv1.CreatePostRequest{
		Title:       "Workout",
		SportTypeId: &sportTypeID,
		Blocks: []*gatewayv1.PostBlockInput{
			{Kind: "text", TextContent: stringPtr("Warm-up")},
			{Kind: "image", FileUrl: stringPtr("https://cdn.example/warm-up.jpg")},
			{Kind: "text", TextContent: stringPtr("Main set")},
		},
	}

	mapped := CreatePostRequestToContent(7, request)

	if mapped.GetAuthorUserId() != 7 {
		t.Fatalf("unexpected author id: %d", mapped.GetAuthorUserId())
	}
	if mapped.SportTypeId == nil || mapped.GetSportTypeId() != 3001 {
		t.Fatalf("unexpected sport type id: %+v", mapped.SportTypeId)
	}
	if len(mapped.GetBlocks()) != 3 {
		t.Fatalf("unexpected blocks count: %d", len(mapped.GetBlocks()))
	}
	if mapped.GetBlocks()[0].GetKind() != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT ||
		mapped.GetBlocks()[0].GetTextContent() != "Warm-up" {
		t.Fatalf("unexpected first block: %+v", mapped.GetBlocks()[0])
	}
	if mapped.GetBlocks()[1].GetKind() != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE ||
		mapped.GetBlocks()[1].GetFileUrl() != "https://cdn.example/warm-up.jpg" {
		t.Fatalf("unexpected second block: %+v", mapped.GetBlocks()[1])
	}
	if mapped.GetBlocks()[2].GetKind() != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT ||
		mapped.GetBlocks()[2].GetTextContent() != "Main set" {
		t.Fatalf("unexpected third block: %+v", mapped.GetBlocks()[2])
	}
}

func TestPostResponseFromContentPreservesBlockOrder(t *testing.T) {
	now := timestamppb.New(time.Date(2026, time.May, 2, 12, 0, 0, 0, time.UTC))
	sportTypeID := int64(3001)
	response, err := PostResponseFromContent(&contentv1.PostResponse{
		Post: &contentv1.Post{
			PostId:        11,
			AuthorUserId:  7,
			Title:         "Workout",
			CreatedAt:     now,
			UpdatedAt:     now,
			CanView:       true,
			LikesCount:    5,
			CommentsCount: 2,
			SportTypeId:   &sportTypeID,
			Blocks: []*contentv1.PostBlock{
				{
					PostBlockId: 101,
					Position:    0,
					Kind:        contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT,
					TextContent: stringPtr("Warm-up"),
				},
				{
					PostBlockId: 102,
					Position:    1,
					Kind:        contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE,
					FileUrl:     stringPtr("https://cdn.example/warm-up.jpg"),
				},
				{
					PostBlockId: 103,
					Position:    2,
					Kind:        contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT,
					TextContent: stringPtr("Main set"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.GetPostId() != 11 || response.GetTrainerId() != 7 || response.GetLikesCount() != 5 || response.GetCommentsCount() != 2 {
		t.Fatalf("unexpected response counters: %+v", response)
	}
	if response.SportTypeId == nil || response.GetSportTypeId() != 3001 {
		t.Fatalf("unexpected sport type id: %+v", response.SportTypeId)
	}
	if len(response.GetBlocks()) != 3 {
		t.Fatalf("unexpected blocks count: %d", len(response.GetBlocks()))
	}
	if response.GetBlocks()[0].GetKind() != "text" || response.GetBlocks()[0].GetTextContent() != "Warm-up" {
		t.Fatalf("unexpected first block: %+v", response.GetBlocks()[0])
	}
	if response.GetBlocks()[1].GetKind() != "image" || response.GetBlocks()[1].GetFileUrl() != "https://cdn.example/warm-up.jpg" {
		t.Fatalf("unexpected second block: %+v", response.GetBlocks()[1])
	}
	if response.GetBlocks()[2].GetKind() != "text" || response.GetBlocks()[2].GetTextContent() != "Main set" {
		t.Fatalf("unexpected third block: %+v", response.GetBlocks()[2])
	}
}

func TestPostMediaUploadResponseFromContent(t *testing.T) {
	response, err := PostMediaUploadResponseFromContent(&contentv1.PostMediaResponse{
		Media: &contentv1.PostMedia{
			FileUrl:     "https://cdn.example/posts/7/run.png",
			Kind:        contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE,
			ContentType: "image/png",
			SizeBytes:   123,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.GetFileUrl() != "https://cdn.example/posts/7/run.png" ||
		response.GetKind() != "image" ||
		response.GetContentType() != "image/png" ||
		response.GetSizeBytes() != 123 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestUpdatePostAndUploadMediaRequestsToContent(t *testing.T) {
	minTierID := int32(2)
	sportTypeID := int32(3)
	title := "Новая тренировка"
	request := &gatewayv1.UpdatePostRequest{
		PostId:           11,
		Title:            &title,
		MinTierId:        &minTierID,
		ClearMinTierId:   true,
		SportTypeId:      &sportTypeID,
		ClearSportTypeId: true,
		ReplaceBlocks:    true,
		Blocks: []*gatewayv1.PostBlockInput{
			{Kind: " video ", FileUrl: stringPtr(" https://cdn/video.mp4 ")},
			{Kind: "unknown", TextContent: stringPtr(" ignored kind ")},
			nil,
		},
	}

	mapped := UpdatePostRequestToContent(7, request)
	if mapped.GetPostId() != 11 || mapped.GetAuthorUserId() != 7 || mapped.GetTitle() != "Новая тренировка" {
		t.Fatalf("unexpected update post request: %+v", mapped)
	}
	if mapped.RequiredSubscriptionLevel == nil || *mapped.RequiredSubscriptionLevel != 2 ||
		!mapped.GetClearRequiredSubscriptionLevel() ||
		mapped.SportTypeId == nil ||
		*mapped.SportTypeId != 3 ||
		!mapped.GetClearSportTypeId() ||
		!mapped.GetReplaceBlocks() {
		t.Fatalf("unexpected update flags: %+v", mapped)
	}
	if len(mapped.GetBlocks()) != 2 ||
		mapped.GetBlocks()[0].GetKind() != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO ||
		mapped.GetBlocks()[0].GetFileUrl() != "https://cdn/video.mp4" ||
		mapped.GetBlocks()[1].GetKind() != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_UNSPECIFIED ||
		mapped.GetBlocks()[1].GetTextContent() != "ignored kind" {
		t.Fatalf("unexpected blocks: %+v", mapped.GetBlocks())
	}

	upload := UploadPostMediaRequestToContent(7, &gatewayv1.UploadPostMediaRequest{
		FileName:    "plan.pdf",
		ContentType: "application/pdf",
		File:        []byte("pdf"),
	})
	if upload.GetAuthorUserId() != 7 || upload.GetFileName() != "plan.pdf" || upload.GetContentType() != "application/pdf" {
		t.Fatalf("unexpected upload request: %+v", upload)
	}
}

func TestSearchPostsRequestToContent(t *testing.T) {
	minTierID := int32(1)
	maxTierID := int32(2)
	viewerLevel := int32(2)

	mapped := SearchPostsRequestToContent(13, &viewerLevel, &gatewayv1.SearchPostsRequest{
		Query:         "темп",
		TrainerIds:    []int32{7, 9},
		SportTypeIds:  []int32{3001},
		BlockKinds:    []string{"image", "document"},
		MinTierId:     &minTierID,
		MaxTierId:     &maxTierID,
		OnlyAvailable: true,
		Limit:         20,
		Offset:        10,
	})

	if mapped.GetQuery() != "темп" ||
		len(mapped.GetAuthorUserIds()) != 2 ||
		mapped.GetAuthorUserIds()[0] != 7 ||
		mapped.GetAuthorUserIds()[1] != 9 ||
		len(mapped.GetSportTypeIds()) != 1 ||
		mapped.GetSportTypeIds()[0] != 3001 ||
		len(mapped.GetBlockKinds()) != 2 ||
		mapped.GetBlockKinds()[0] != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE ||
		mapped.GetBlockKinds()[1] != contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT ||
		mapped.MinRequiredSubscriptionLevel == nil ||
		*mapped.MinRequiredSubscriptionLevel != 1 ||
		mapped.MaxRequiredSubscriptionLevel == nil ||
		*mapped.MaxRequiredSubscriptionLevel != 2 ||
		!mapped.GetOnlyAvailable() ||
		mapped.GetViewerUserId() != 13 ||
		mapped.ViewerSubscriptionLevel == nil ||
		*mapped.ViewerSubscriptionLevel != 2 ||
		mapped.GetLimit() != 20 ||
		mapped.GetOffset() != 10 {
		t.Fatalf("unexpected search request: %+v", mapped)
	}
}

func TestPostCollectionResponsesFromContent(t *testing.T) {
	now := timestamppb.New(time.Date(2026, time.May, 2, 12, 0, 0, 0, time.UTC))
	tierID := int32(1)
	sportTypeID := int64(3001)
	summary := &contentv1.PostSummary{
		PostId:                    11,
		AuthorUserId:              7,
		Title:                     "Темповая работа",
		RequiredSubscriptionLevel: &tierID,
		SportTypeId:               &sportTypeID,
		CreatedAt:                 now,
		CanView:                   true,
		LikesCount:                4,
		IsLiked:                   true,
		CommentsCount:             1,
	}

	search, err := SearchPostsResponseFromContent(&contentv1.SearchPostsResponse{Posts: []*contentv1.PostSummary{summary}})
	if err != nil {
		t.Fatalf("unexpected search response error: %v", err)
	}
	if len(search.GetPosts()) != 1 ||
		search.GetPosts()[0].GetPostId() != 11 ||
		search.GetPosts()[0].GetSportTypeId() != 3001 ||
		search.GetPosts()[0].GetLikesCount() != 4 ||
		!search.GetPosts()[0].GetIsLiked() {
		t.Fatalf("unexpected search response: %+v", search)
	}

	profilePosts, err := ProfilePostsResponseFromContent(7, &contentv1.ListAuthorPostsResponse{Posts: []*contentv1.PostSummary{summary}})
	if err != nil {
		t.Fatalf("unexpected profile posts response error: %v", err)
	}
	if profilePosts.GetUserId() != 7 || len(profilePosts.GetPosts()) != 1 {
		t.Fatalf("unexpected profile posts response: %+v", profilePosts)
	}

	emptySearch, err := SearchPostsResponseFromContent(nil)
	if err != nil || len(emptySearch.GetPosts()) != 0 {
		t.Fatalf("unexpected empty search response: %+v err=%v", emptySearch, err)
	}
	emptyProfilePosts, err := ProfilePostsResponseFromContent(7, nil)
	if err != nil || len(emptyProfilePosts.GetPosts()) != 0 {
		t.Fatalf("unexpected empty profile posts response: %+v err=%v", emptyProfilePosts, err)
	}
}

func TestPostLikeResponseFromContent(t *testing.T) {
	response, err := PostLikeResponseFromContent(&contentv1.PostLikeStateResponse{
		State: &contentv1.PostLikeState{
			PostId:     11,
			LikesCount: 5,
			IsLiked:    true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected like response error: %v", err)
	}
	if response.GetPostId() != 11 || response.GetLikesCount() != 5 || !response.GetIsLiked() {
		t.Fatalf("unexpected like response: %+v", response)
	}
	if _, err := PostLikeResponseFromContent(nil); err == nil {
		t.Fatal("expected nil like state error")
	}
}

func stringPtr(value string) *string {
	return &value
}
