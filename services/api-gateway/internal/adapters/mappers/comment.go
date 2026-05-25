package mappers

import (
	"fmt"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func CreateCommentRequestToContent(
	authorUserID int64,
	viewerSubscriptionLevel *int32,
	request *gatewayv1.CreateCommentRequest,
) *contentv1.CreateCommentRequest {
	return &contentv1.CreateCommentRequest{
		PostId:                  int32ToInt64(request.GetPostId()),
		AuthorUserId:            authorUserID,
		ViewerSubscriptionLevel: viewerSubscriptionLevel,
		Body:                    request.GetBody(),
	}
}

func ListCommentsRequestToContent(
	viewerUserID int64,
	viewerSubscriptionLevel *int32,
	request *gatewayv1.ListCommentsRequest,
) *contentv1.ListCommentsRequest {
	return &contentv1.ListCommentsRequest{
		PostId:                  int32ToInt64(request.GetPostId()),
		ViewerUserId:            viewerUserID,
		ViewerSubscriptionLevel: viewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func ListPostLikesRequestToContent(
	viewerUserID int64,
	viewerSubscriptionLevel *int32,
	request *gatewayv1.ListPostLikesRequest,
) *contentv1.ListPostLikesRequest {
	return &contentv1.ListPostLikesRequest{
		PostId:                  int32ToInt64(request.GetPostId()),
		ViewerUserId:            viewerUserID,
		ViewerSubscriptionLevel: viewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func ListPostLikesResponseFromContent(response *contentv1.ListPostLikesResponse) (*gatewayv1.ListPostLikesResponse, error) {
	likes := make([]*gatewayv1.PostLike, 0)
	if response != nil {
		likes = make([]*gatewayv1.PostLike, 0, len(response.GetLikes()))
		for _, like := range response.GetLikes() {
			if like == nil {
				continue
			}

			userID, err := int64ToInt32("content.post_like.user_id", like.GetUserId())
			if err != nil {
				return nil, err
			}

			likes = append(likes, &gatewayv1.PostLike{
				UserId:    userID,
				CreatedAt: like.GetCreatedAt(),
			})
		}
	}

	return &gatewayv1.ListPostLikesResponse{Likes: likes}, nil
}

func CommentResponseFromContent(response *contentv1.CommentResponse) (*gatewayv1.CommentResponse, error) {
	if response == nil || response.GetComment() == nil {
		return nil, fmt.Errorf("comment is required")
	}

	comment, err := commentFromContent(response.GetComment())
	if err != nil {
		return nil, err
	}

	return &gatewayv1.CommentResponse{Comment: comment}, nil
}

func ListCommentsResponseFromContent(response *contentv1.ListCommentsResponse) (*gatewayv1.ListCommentsResponse, error) {
	comments := make([]*gatewayv1.Comment, 0)
	if response != nil {
		comments = make([]*gatewayv1.Comment, 0, len(response.GetComments()))
		for _, comment := range response.GetComments() {
			if comment == nil {
				continue
			}

			mappedComment, err := commentFromContent(comment)
			if err != nil {
				return nil, err
			}

			comments = append(comments, mappedComment)
		}
	}

	return &gatewayv1.ListCommentsResponse{Comments: comments}, nil
}

func commentFromContent(comment *contentv1.Comment) (*gatewayv1.Comment, error) {
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
