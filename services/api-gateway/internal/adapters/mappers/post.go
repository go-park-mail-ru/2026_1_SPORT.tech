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
		Blocks:                    postBlockInputsToContent(request.GetBlocks()),
	}
}

func UpdatePostRequestToContent(authorUserID int64, request *gatewayv1.UpdatePostRequest) *contentv1.UpdatePostRequest {
	return &contentv1.UpdatePostRequest{
		PostId:                         int32ToInt64(request.GetPostId()),
		AuthorUserId:                   authorUserID,
		Title:                          request.Title,
		RequiredSubscriptionLevel:      request.MinTierId,
		ClearRequiredSubscriptionLevel: request.GetClearMinTierId(),
		Blocks:                         postBlockInputsToContent(request.GetBlocks()),
		ReplaceBlocks:                  request.GetReplaceBlocks(),
	}
}

func UploadPostMediaRequestToContent(authorUserID int64, request *gatewayv1.UploadPostMediaRequest) *contentv1.UploadPostMediaRequest {
	return &contentv1.UploadPostMediaRequest{
		AuthorUserId: authorUserID,
		File:         request.GetFile(),
		FileName:     request.GetFileName(),
		ContentType:  request.GetContentType(),
	}
}

func PostResponseFromContent(response *contentv1.PostResponse) (*gatewayv1.PostResponse, error) {
	if response == nil || response.GetPost() == nil {
		return nil, fmt.Errorf("post is required")
	}

	return postResponseFromContentPost(response.GetPost())
}

func PostMediaUploadResponseFromContent(response *contentv1.PostMediaResponse) (*gatewayv1.PostMediaUploadResponse, error) {
	if response == nil || response.GetMedia() == nil {
		return nil, fmt.Errorf("post media is required")
	}

	sizeBytes, err := int64ToInt32("content.post_media.size_bytes", response.GetMedia().GetSizeBytes())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PostMediaUploadResponse{
		FileUrl:     response.GetMedia().GetFileUrl(),
		Kind:        blockKindFromContent(response.GetMedia().GetKind()),
		ContentType: response.GetMedia().GetContentType(),
		SizeBytes:   sizeBytes,
	}, nil
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

	commentsCount, err := int64ToInt32("content.post.comments_count", post.GetCommentsCount())
	if err != nil {
		return nil, err
	}

	blocks, err := postBlocksFromContent(post.GetBlocks())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PostResponse{
		PostId:        postID,
		TrainerId:     trainerID,
		MinTierId:     post.RequiredSubscriptionLevel,
		Title:         post.GetTitle(),
		CreatedAt:     post.GetCreatedAt(),
		UpdatedAt:     post.GetUpdatedAt(),
		LikesCount:    likesCount,
		IsLiked:       post.GetIsLiked(),
		Blocks:        blocks,
		CanView:       post.GetCanView(),
		CommentsCount: commentsCount,
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

	commentsCount, err := int64ToInt32("content.post_summary.comments_count", post.GetCommentsCount())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PostListItem{
		PostId:        postID,
		TrainerId:     trainerID,
		MinTierId:     post.RequiredSubscriptionLevel,
		Title:         post.GetTitle(),
		CreatedAt:     post.GetCreatedAt(),
		CanView:       post.GetCanView(),
		LikesCount:    likesCount,
		IsLiked:       post.GetIsLiked(),
		CommentsCount: commentsCount,
	}, nil
}

func postBlockInputsToContent(blocks []*gatewayv1.PostBlockInput) []*contentv1.PostBlockInput {
	result := make([]*contentv1.PostBlockInput, 0, len(blocks))
	for _, block := range blocks {
		if block == nil {
			continue
		}

		result = append(result, &contentv1.PostBlockInput{
			Kind:        blockKindToContent(block.GetKind()),
			TextContent: trimOptionalString(block.TextContent),
			FileUrl:     trimOptionalString(block.FileUrl),
		})
	}

	return result
}

func postBlocksFromContent(blocks []*contentv1.PostBlock) ([]*gatewayv1.PostBlock, error) {
	result := make([]*gatewayv1.PostBlock, 0, len(blocks))
	for _, block := range blocks {
		postBlockID, err := int64ToInt32("content.post_block.post_block_id", block.GetPostBlockId())
		if err != nil {
			return nil, err
		}

		result = append(result, &gatewayv1.PostBlock{
			PostBlockId: postBlockID,
			Position:    block.GetPosition(),
			Kind:        blockKindFromContent(block.GetKind()),
			TextContent: block.TextContent,
			FileUrl:     block.FileUrl,
		})
	}

	return result, nil
}

func blockKindToContent(kind string) contentv1.ContentBlockKind {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "text":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT
	case "image":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE
	case "video":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO
	case "document":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT
	default:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_UNSPECIFIED
	}
}

func blockKindFromContent(kind contentv1.ContentBlockKind) string {
	switch kind {
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT:
		return "text"
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE:
		return "image"
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO:
		return "video"
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT:
		return "document"
	default:
		return ""
	}
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
