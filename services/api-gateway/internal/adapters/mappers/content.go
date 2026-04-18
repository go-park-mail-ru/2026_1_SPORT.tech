package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func ListAuthorPostsRequestToContent(request *gatewayv1.ListAuthorPostsRequest) *contentv1.ListAuthorPostsRequest {
	return &contentv1.ListAuthorPostsRequest{
		AuthorUserId:            request.GetAuthorUserId(),
		ViewerUserId:            request.GetViewerUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func CreatePostRequestToContent(request *gatewayv1.CreatePostRequest) *contentv1.CreatePostRequest {
	return &contentv1.CreatePostRequest{
		AuthorUserId:              request.GetAuthorUserId(),
		Title:                     request.GetTitle(),
		RequiredSubscriptionLevel: request.RequiredSubscriptionLevel,
		Blocks:                    postBlockInputsToContent(request.GetBlocks()),
	}
}

func GetPostRequestToContent(request *gatewayv1.GetPostRequest) *contentv1.GetPostRequest {
	return &contentv1.GetPostRequest{
		PostId:                  request.GetPostId(),
		ViewerUserId:            request.GetViewerUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func UpdatePostRequestToContent(request *gatewayv1.UpdatePostRequest) *contentv1.UpdatePostRequest {
	return &contentv1.UpdatePostRequest{
		PostId:                         request.GetPostId(),
		AuthorUserId:                   request.GetAuthorUserId(),
		Title:                          request.Title,
		RequiredSubscriptionLevel:      request.RequiredSubscriptionLevel,
		ClearRequiredSubscriptionLevel: request.GetClearRequiredSubscriptionLevel(),
		Blocks:                         postBlockInputsToContent(request.GetBlocks()),
		ReplaceBlocks:                  request.GetReplaceBlocks(),
	}
}

func DeletePostRequestToContent(request *gatewayv1.DeletePostRequest) *contentv1.DeletePostRequest {
	return &contentv1.DeletePostRequest{
		PostId:       request.GetPostId(),
		AuthorUserId: request.GetAuthorUserId(),
	}
}

func LikePostRequestToContent(request *gatewayv1.LikePostRequest) *contentv1.LikePostRequest {
	return &contentv1.LikePostRequest{
		PostId:                  request.GetPostId(),
		UserId:                  request.GetUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func UnlikePostRequestToContent(request *gatewayv1.UnlikePostRequest) *contentv1.UnlikePostRequest {
	return &contentv1.UnlikePostRequest{
		PostId:                  request.GetPostId(),
		UserId:                  request.GetUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func CreateCommentRequestToContent(request *gatewayv1.CreateCommentRequest) *contentv1.CreateCommentRequest {
	return &contentv1.CreateCommentRequest{
		PostId:                  request.GetPostId(),
		AuthorUserId:            request.GetAuthorUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Body:                    request.GetBody(),
	}
}

func ListCommentsRequestToContent(request *gatewayv1.ListCommentsRequest) *contentv1.ListCommentsRequest {
	return &contentv1.ListCommentsRequest{
		PostId:                  request.GetPostId(),
		ViewerUserId:            request.GetViewerUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func PostResponseFromContent(response *contentv1.PostResponse) *gatewayv1.PostResponse {
	if response == nil {
		return nil
	}

	return &gatewayv1.PostResponse{Post: postFromContent(response.GetPost())}
}

func ListAuthorPostsResponseFromContent(response *contentv1.ListAuthorPostsResponse) *gatewayv1.ListAuthorPostsResponse {
	if response == nil {
		return nil
	}

	posts := make([]*gatewayv1.PostSummary, 0, len(response.GetPosts()))
	for _, post := range response.GetPosts() {
		posts = append(posts, postSummaryFromContent(post))
	}

	return &gatewayv1.ListAuthorPostsResponse{Posts: posts}
}

func PostLikeStateResponseFromContent(response *contentv1.PostLikeStateResponse) *gatewayv1.PostLikeStateResponse {
	if response == nil {
		return nil
	}

	state := response.GetState()
	if state == nil {
		return &gatewayv1.PostLikeStateResponse{}
	}

	return &gatewayv1.PostLikeStateResponse{
		State: &gatewayv1.PostLikeState{
			PostId:     state.GetPostId(),
			LikesCount: state.GetLikesCount(),
			IsLiked:    state.GetIsLiked(),
		},
	}
}

func CommentResponseFromContent(response *contentv1.CommentResponse) *gatewayv1.CommentResponse {
	if response == nil {
		return nil
	}

	return &gatewayv1.CommentResponse{Comment: commentFromContent(response.GetComment())}
}

func ListCommentsResponseFromContent(response *contentv1.ListCommentsResponse) *gatewayv1.ListCommentsResponse {
	if response == nil {
		return nil
	}

	comments := make([]*gatewayv1.Comment, 0, len(response.GetComments()))
	for _, comment := range response.GetComments() {
		comments = append(comments, commentFromContent(comment))
	}

	return &gatewayv1.ListCommentsResponse{Comments: comments}
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

func postFromContent(post *contentv1.Post) *gatewayv1.Post {
	if post == nil {
		return nil
	}

	blocks := make([]*gatewayv1.PostBlock, 0, len(post.GetBlocks()))
	for _, block := range post.GetBlocks() {
		blocks = append(blocks, &gatewayv1.PostBlock{
			PostBlockId: block.GetPostBlockId(),
			Position:    block.GetPosition(),
			Kind:        gatewayv1.ContentBlockKind(block.GetKind()),
			TextContent: block.TextContent,
			FileUrl:     block.FileUrl,
		})
	}

	return &gatewayv1.Post{
		PostId:                    post.GetPostId(),
		AuthorUserId:              post.GetAuthorUserId(),
		Title:                     post.GetTitle(),
		RequiredSubscriptionLevel: post.RequiredSubscriptionLevel,
		CreatedAt:                 post.GetCreatedAt(),
		UpdatedAt:                 post.GetUpdatedAt(),
		CanView:                   post.GetCanView(),
		LikesCount:                post.GetLikesCount(),
		IsLiked:                   post.GetIsLiked(),
		CommentsCount:             post.GetCommentsCount(),
		Blocks:                    blocks,
	}
}

func postSummaryFromContent(post *contentv1.PostSummary) *gatewayv1.PostSummary {
	if post == nil {
		return nil
	}

	return &gatewayv1.PostSummary{
		PostId:                    post.GetPostId(),
		AuthorUserId:              post.GetAuthorUserId(),
		Title:                     post.GetTitle(),
		RequiredSubscriptionLevel: post.RequiredSubscriptionLevel,
		CreatedAt:                 post.GetCreatedAt(),
		CanView:                   post.GetCanView(),
		LikesCount:                post.GetLikesCount(),
		IsLiked:                   post.GetIsLiked(),
		CommentsCount:             post.GetCommentsCount(),
	}
}

func commentFromContent(comment *contentv1.Comment) *gatewayv1.Comment {
	if comment == nil {
		return nil
	}

	return &gatewayv1.Comment{
		CommentId:    comment.GetCommentId(),
		PostId:       comment.GetPostId(),
		AuthorUserId: comment.GetAuthorUserId(),
		Body:         comment.GetBody(),
		CreatedAt:    comment.GetCreatedAt(),
		UpdatedAt:    comment.GetUpdatedAt(),
	}
}
