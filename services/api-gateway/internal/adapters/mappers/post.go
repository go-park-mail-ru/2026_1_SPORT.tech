package mappers

import (
	"fmt"
	"strings"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func CreatePostRequestToContent(authorUserID int64, request *gatewayv1.CreatePostRequest) *contentv1.CreatePostRequest {
	return &contentv1.CreatePostRequest{
		AuthorUserId:              authorUserID,
		Title:                     request.GetTitle(),
		RequiredSubscriptionLevel: request.MinTierId,
		Blocks:                    postBlocksToContent(request.GetTextContent(), request.GetAttachments()),
	}
}

func UpdatePostRequestToContent(authorUserID int64, request *gatewayv1.UpdatePostRequest) *contentv1.UpdatePostRequest {
	replaceBlocks := request.TextContent != nil || len(request.GetAttachments()) > 0

	return &contentv1.UpdatePostRequest{
		PostId:                         int32ToInt64(request.GetPostId()),
		AuthorUserId:                   authorUserID,
		Title:                          request.Title,
		RequiredSubscriptionLevel:      request.MinTierId,
		ClearRequiredSubscriptionLevel: false,
		Blocks:                         postBlocksToContent(request.GetTextContent(), request.GetAttachments()),
		ReplaceBlocks:                  replaceBlocks,
	}
}

func PostResponseFromContent(response *contentv1.PostResponse) (*gatewayv1.PostResponse, error) {
	if response == nil || response.GetPost() == nil {
		return nil, fmt.Errorf("post is required")
	}

	return postResponseFromContentPost(response.GetPost())
}

func ProfilePostsResponseFromContent(userID int32, response *contentv1.ListAuthorPostsResponse) (*gatewayv1.ProfilePostsResponse, error) {
	posts := make([]*gatewayv1.PostListItem, 0)
	if response != nil {
		posts = make([]*gatewayv1.PostListItem, 0, len(response.GetPosts()))
		for _, post := range response.GetPosts() {
			mappedPost, err := postListItemFromContent(post)
			if err != nil {
				return nil, err
			}

			posts = append(posts, mappedPost)
		}
	}

	return &gatewayv1.ProfilePostsResponse{
		UserId: userID,
		Posts:  posts,
	}, nil
}

func PostLikeResponseFromContent(response *contentv1.PostLikeStateResponse) (*gatewayv1.PostLikeResponse, error) {
	if response == nil || response.GetState() == nil {
		return nil, fmt.Errorf("post like state is required")
	}

	postID, err := int64ToInt32("content.post_like_state.post_id", response.GetState().GetPostId())
	if err != nil {
		return nil, err
	}

	likesCount, err := int64ToInt32("content.post_like_state.likes_count", response.GetState().GetLikesCount())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PostLikeResponse{
		PostId:     postID,
		LikesCount: likesCount,
		IsLiked:    response.GetState().GetIsLiked(),
	}, nil
}

func postResponseFromContentPost(post *contentv1.Post) (*gatewayv1.PostResponse, error) {
	postID, err := int64ToInt32("content.post.post_id", post.GetPostId())
	if err != nil {
		return nil, err
	}

	trainerID, err := int64ToInt32("content.post.author_user_id", post.GetAuthorUserId())
	if err != nil {
		return nil, err
	}

	likesCount, err := int64ToInt32("content.post.likes_count", post.GetLikesCount())
	if err != nil {
		return nil, err
	}

	textContent, attachments, err := flattenPostBlocks(post.GetBlocks())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PostResponse{
		PostId:      postID,
		TrainerId:   trainerID,
		MinTierId:   post.RequiredSubscriptionLevel,
		Title:       post.GetTitle(),
		TextContent: textContent,
		CreatedAt:   post.GetCreatedAt(),
		UpdatedAt:   post.GetUpdatedAt(),
		LikesCount:  likesCount,
		IsLiked:     post.GetIsLiked(),
		Attachments: attachments,
	}, nil
}

func postListItemFromContent(post *contentv1.PostSummary) (*gatewayv1.PostListItem, error) {
	postID, err := int64ToInt32("content.post_summary.post_id", post.GetPostId())
	if err != nil {
		return nil, err
	}

	trainerID, err := int64ToInt32("content.post_summary.author_user_id", post.GetAuthorUserId())
	if err != nil {
		return nil, err
	}

	likesCount, err := int64ToInt32("content.post_summary.likes_count", post.GetLikesCount())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PostListItem{
		PostId:     postID,
		TrainerId:  trainerID,
		MinTierId:  post.RequiredSubscriptionLevel,
		Title:      post.GetTitle(),
		CreatedAt:  post.GetCreatedAt(),
		CanView:    post.GetCanView(),
		LikesCount: likesCount,
		IsLiked:    post.GetIsLiked(),
	}, nil
}

func postBlocksToContent(textContent string, attachments []*gatewayv1.CreatePostAttachmentRequest) []*contentv1.PostBlockInput {
	blocks := make([]*contentv1.PostBlockInput, 0, 1+len(attachments))
	if trimmedText := strings.TrimSpace(textContent); trimmedText != "" {
		blocks = append(blocks, &contentv1.PostBlockInput{
			Kind:        contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT,
			TextContent: &trimmedText,
		})
	}

	for _, attachment := range attachments {
		if attachment == nil {
			continue
		}

		kind := attachmentKindToContent(attachment.GetKind())
		fileURL := attachment.GetFileUrl()
		blocks = append(blocks, &contentv1.PostBlockInput{
			Kind:    kind,
			FileUrl: &fileURL,
		})
	}

	return blocks
}

func flattenPostBlocks(blocks []*contentv1.PostBlock) (string, []*gatewayv1.PostAttachment, error) {
	textParts := make([]string, 0)
	attachments := make([]*gatewayv1.PostAttachment, 0)

	for _, block := range blocks {
		switch block.GetKind() {
		case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT:
			if block.TextContent != nil {
				textParts = append(textParts, block.GetTextContent())
			}
		case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE,
			contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO,
			contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT:
			postAttachmentID, err := int64ToInt32("content.post_block.post_block_id", block.GetPostBlockId())
			if err != nil {
				return "", nil, err
			}

			attachments = append(attachments, &gatewayv1.PostAttachment{
				PostAttachmentId: postAttachmentID,
				Kind:             attachmentKindFromContent(block.GetKind()),
				FileUrl:          block.GetFileUrl(),
			})
		}
	}

	return strings.Join(textParts, "\n\n"), attachments, nil
}

func attachmentKindToContent(kind string) contentv1.ContentBlockKind {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "video":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO
	case "document":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT
	default:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE
	}
}

func attachmentKindFromContent(kind contentv1.ContentBlockKind) string {
	switch kind {
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO:
		return "video"
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT:
		return "document"
	default:
		return "image"
	}
}
