package mappers

import (
	"testing"
	"time"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCreatePostRequestToContentPreservesBlockOrder(t *testing.T) {
	request := &gatewayv1.CreatePostRequest{
		Title: "Workout",
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

func stringPtr(value string) *string {
	return &value
}
