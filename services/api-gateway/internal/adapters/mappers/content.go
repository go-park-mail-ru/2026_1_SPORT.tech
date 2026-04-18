package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func ListAuthorPostsRequestToContent(request *gatewayv1.ListAuthorPostsRequest) *contentv1.ListAuthorPostsRequest {
	return &contentv1.ListAuthorPostsRequest{
		AuthorUserId:            int32ToInt64(request.GetAuthorUserId()),
		ViewerUserId:            int32ToInt64(request.GetViewerUserId()),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func CreatePostRequestToContent(request *gatewayv1.CreatePostRequest) *contentv1.CreatePostRequest {
	return &contentv1.CreatePostRequest{
		AuthorUserId:              int32ToInt64(request.GetAuthorUserId()),
		Title:                     request.GetTitle(),
		RequiredSubscriptionLevel: request.RequiredSubscriptionLevel,
		Blocks:                    postBlockInputsToContent(request.GetBlocks()),
	}
}

func GetPostRequestToContent(request *gatewayv1.GetPostRequest) *contentv1.GetPostRequest {
	return &contentv1.GetPostRequest{
		PostId:                  int32ToInt64(request.GetPostId()),
		ViewerUserId:            int32ToInt64(request.GetViewerUserId()),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func UpdatePostRequestToContent(request *gatewayv1.UpdatePostRequest) *contentv1.UpdatePostRequest {
	return &contentv1.UpdatePostRequest{
		PostId:                         int32ToInt64(request.GetPostId()),
		AuthorUserId:                   int32ToInt64(request.GetAuthorUserId()),
		Title:                          request.Title,
		RequiredSubscriptionLevel:      request.RequiredSubscriptionLevel,
		ClearRequiredSubscriptionLevel: request.GetClearRequiredSubscriptionLevel(),
		Blocks:                         postBlockInputsToContent(request.GetBlocks()),
		ReplaceBlocks:                  request.GetReplaceBlocks(),
	}
}

func DeletePostRequestToContent(request *gatewayv1.DeletePostRequest) *contentv1.DeletePostRequest {
	return &contentv1.DeletePostRequest{
		PostId:       int32ToInt64(request.GetPostId()),
		AuthorUserId: int32ToInt64(request.GetAuthorUserId()),
	}
}

func LikePostRequestToContent(request *gatewayv1.LikePostRequest) *contentv1.LikePostRequest {
	return &contentv1.LikePostRequest{
		PostId:                  int32ToInt64(request.GetPostId()),
		UserId:                  int32ToInt64(request.GetUserId()),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func UnlikePostRequestToContent(request *gatewayv1.UnlikePostRequest) *contentv1.UnlikePostRequest {
	return &contentv1.UnlikePostRequest{
		PostId:                  int32ToInt64(request.GetPostId()),
		UserId:                  int32ToInt64(request.GetUserId()),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func CreateCommentRequestToContent(request *gatewayv1.CreateCommentRequest) *contentv1.CreateCommentRequest {
	return &contentv1.CreateCommentRequest{
		PostId:                  int32ToInt64(request.GetPostId()),
		AuthorUserId:            int32ToInt64(request.GetAuthorUserId()),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Body:                    request.GetBody(),
	}
}

func ListCommentsRequestToContent(request *gatewayv1.ListCommentsRequest) *contentv1.ListCommentsRequest {
	return &contentv1.ListCommentsRequest{
		PostId:                  int32ToInt64(request.GetPostId()),
		ViewerUserId:            int32ToInt64(request.GetViewerUserId()),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func PostResponseFromContent(response *contentv1.PostResponse) (*gatewayv1.PostResponse, error) {
	if response == nil {
		return nil, nil
	}

	post, err := postFromContent(response.GetPost())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PostResponse{Post: post}, nil
}

func ListAuthorPostsResponseFromContent(response *contentv1.ListAuthorPostsResponse) (*gatewayv1.ListAuthorPostsResponse, error) {
	if response == nil {
		return nil, nil
	}

	posts := make([]*gatewayv1.PostSummary, 0, len(response.GetPosts()))
	for _, post := range response.GetPosts() {
		mappedPost, err := postSummaryFromContent(post)
		if err != nil {
			return nil, err
		}

		posts = append(posts, mappedPost)
	}

	return &gatewayv1.ListAuthorPostsResponse{Posts: posts}, nil
}

func PostLikeStateResponseFromContent(response *contentv1.PostLikeStateResponse) (*gatewayv1.PostLikeStateResponse, error) {
	if response == nil {
		return nil, nil
	}

	state := response.GetState()
	if state == nil {
		return &gatewayv1.PostLikeStateResponse{}, nil
	}

	postID, err := int64ToInt32("content.post_like_state.post_id", state.GetPostId())
	if err != nil {
		return nil, err
	}

	likesCount, err := int64ToInt32("content.post_like_state.likes_count", state.GetLikesCount())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.PostLikeStateResponse{
		State: &gatewayv1.PostLikeState{
			PostId:     postID,
			LikesCount: likesCount,
			IsLiked:    state.GetIsLiked(),
		},
	}, nil
}

func CommentResponseFromContent(response *contentv1.CommentResponse) (*gatewayv1.CommentResponse, error) {
	if response == nil {
		return nil, nil
	}

	comment, err := commentFromContent(response.GetComment())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.CommentResponse{Comment: comment}, nil
}

func ListCommentsResponseFromContent(response *contentv1.ListCommentsResponse) (*gatewayv1.ListCommentsResponse, error) {
	if response == nil {
		return nil, nil
	}

	comments := make([]*gatewayv1.Comment, 0, len(response.GetComments()))
	for _, comment := range response.GetComments() {
		mappedComment, err := commentFromContent(comment)
		if err != nil {
			return nil, err
		}

		comments = append(comments, mappedComment)
	}

	return &gatewayv1.ListCommentsResponse{Comments: comments}, nil
}

func postBlockInputsToContent(blocks []*gatewayv1.PostBlockInput) []*contentv1.PostBlockInput {
	result := make([]*contentv1.PostBlockInput, 0, len(blocks))
	for _, block := range blocks {
		result = append(result, &contentv1.PostBlockInput{
			Kind:        contentv1.ContentBlockKind(block.GetKind()),
			TextContent: block.TextContent,
			FileUrl:     block.FileUrl,
		})
	}

	return result
}

func postFromContent(post *contentv1.Post) (*gatewayv1.Post, error) {
	if post == nil {
		return nil, nil
	}

	blocks := make([]*gatewayv1.PostBlock, 0, len(post.GetBlocks()))
	for _, block := range post.GetBlocks() {
		postBlockID, err := int64ToInt32("content.post_block.post_block_id", block.GetPostBlockId())
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, &gatewayv1.PostBlock{
			PostBlockId: postBlockID,
			Position:    block.GetPosition(),
			Kind:        gatewayv1.ContentBlockKind(block.GetKind()),
			TextContent: block.TextContent,
			FileUrl:     block.FileUrl,
		})
	}

	postID, err := int64ToInt32("content.post.post_id", post.GetPostId())
	if err != nil {
		return nil, err
	}

	authorUserID, err := int64ToInt32("content.post.author_user_id", post.GetAuthorUserId())
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

	return &gatewayv1.Post{
		PostId:                    postID,
		AuthorUserId:              authorUserID,
		Title:                     post.GetTitle(),
		RequiredSubscriptionLevel: post.RequiredSubscriptionLevel,
		CreatedAt:                 post.GetCreatedAt(),
		UpdatedAt:                 post.GetUpdatedAt(),
		CanView:                   post.GetCanView(),
		LikesCount:                likesCount,
		IsLiked:                   post.GetIsLiked(),
		CommentsCount:             commentsCount,
		Blocks:                    blocks,
	}, nil
}

func postSummaryFromContent(post *contentv1.PostSummary) (*gatewayv1.PostSummary, error) {
	if post == nil {
		return nil, nil
	}

	postID, err := int64ToInt32("content.post_summary.post_id", post.GetPostId())
	if err != nil {
		return nil, err
	}

	authorUserID, err := int64ToInt32("content.post_summary.author_user_id", post.GetAuthorUserId())
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

	return &gatewayv1.PostSummary{
		PostId:                    postID,
		AuthorUserId:              authorUserID,
		Title:                     post.GetTitle(),
		RequiredSubscriptionLevel: post.RequiredSubscriptionLevel,
		CreatedAt:                 post.GetCreatedAt(),
		CanView:                   post.GetCanView(),
		LikesCount:                likesCount,
		IsLiked:                   post.GetIsLiked(),
		CommentsCount:             commentsCount,
	}, nil
}

func commentFromContent(comment *contentv1.Comment) (*gatewayv1.Comment, error) {
	if comment == nil {
		return nil, nil
	}

	commentID, err := int64ToInt32("content.comment.comment_id", comment.GetCommentId())
	if err != nil {
		return nil, err
	}

	postID, err := int64ToInt32("content.comment.post_id", comment.GetPostId())
	if err != nil {
		return nil, err
	}

	authorUserID, err := int64ToInt32("content.comment.author_user_id", comment.GetAuthorUserId())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.Comment{
		CommentId:    commentID,
		PostId:       postID,
		AuthorUserId: authorUserID,
		Body:         comment.GetBody(),
		CreatedAt:    comment.GetCreatedAt(),
		UpdatedAt:    comment.GetUpdatedAt(),
	}, nil
}
