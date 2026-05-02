package mappers

import (
	"errors"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ListAuthorPostsRequestToQuery(request *contentv1.ListAuthorPostsRequest) usecase.ListAuthorPostsQuery {
	return usecase.ListAuthorPostsQuery{
		AuthorUserID:            request.GetAuthorUserId(),
		ViewerUserID:            request.GetViewerUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func SearchPostsRequestToQuery(request *contentv1.SearchPostsRequest) usecase.SearchPostsQuery {
	return usecase.SearchPostsQuery{
		Query:                        request.GetQuery(),
		AuthorUserIDs:                request.GetAuthorUserIds(),
		BlockKinds:                   blockKindsFromProto(request.GetBlockKinds()),
		MinRequiredSubscriptionLevel: request.MinRequiredSubscriptionLevel,
		MaxRequiredSubscriptionLevel: request.MaxRequiredSubscriptionLevel,
		OnlyAvailable:                request.GetOnlyAvailable(),
		ViewerUserID:                 request.GetViewerUserId(),
		ViewerSubscriptionLevel:      request.ViewerSubscriptionLevel,
		Limit:                        request.GetLimit(),
		Offset:                       request.GetOffset(),
	}
}

func CreatePostRequestToCommand(request *contentv1.CreatePostRequest) usecase.CreatePostCommand {
	return usecase.CreatePostCommand{
		AuthorUserID:              request.GetAuthorUserId(),
		Title:                     request.GetTitle(),
		RequiredSubscriptionLevel: request.RequiredSubscriptionLevel,
		Blocks:                    postBlockInputsFromProto(request.GetBlocks()),
	}
}

func UploadPostMediaRequestToCommand(request *contentv1.UploadPostMediaRequest) usecase.UploadPostMediaCommand {
	return usecase.UploadPostMediaCommand{
		AuthorUserID: request.GetAuthorUserId(),
		FileName:     request.GetFileName(),
		ContentType:  request.GetContentType(),
		Content:      request.GetFile(),
	}
}

func GetPostRequestToQuery(request *contentv1.GetPostRequest) usecase.GetPostQuery {
	return usecase.GetPostQuery{
		PostID:                  request.GetPostId(),
		ViewerUserID:            request.GetViewerUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func UpdatePostRequestToCommand(request *contentv1.UpdatePostRequest) usecase.UpdatePostCommand {
	return usecase.UpdatePostCommand{
		PostID:                         request.GetPostId(),
		AuthorUserID:                   request.GetAuthorUserId(),
		Title:                          request.Title,
		RequiredSubscriptionLevel:      request.RequiredSubscriptionLevel,
		ClearRequiredSubscriptionLevel: request.GetClearRequiredSubscriptionLevel(),
		Blocks:                         postBlockInputsFromProto(request.GetBlocks()),
		ReplaceBlocks:                  request.GetReplaceBlocks(),
	}
}

func DeletePostRequestToCommand(request *contentv1.DeletePostRequest) usecase.DeletePostCommand {
	return usecase.DeletePostCommand{
		PostID:       request.GetPostId(),
		AuthorUserID: request.GetAuthorUserId(),
	}
}

func LikePostRequestToCommand(request *contentv1.LikePostRequest) usecase.LikePostCommand {
	return usecase.LikePostCommand{
		PostID:                  request.GetPostId(),
		UserID:                  request.GetUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func UnlikePostRequestToCommand(request *contentv1.UnlikePostRequest) usecase.LikePostCommand {
	return usecase.LikePostCommand{
		PostID:                  request.GetPostId(),
		UserID:                  request.GetUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
	}
}

func CreateCommentRequestToCommand(request *contentv1.CreateCommentRequest) usecase.CreateCommentCommand {
	return usecase.CreateCommentCommand{
		PostID:                  request.GetPostId(),
		AuthorUserID:            request.GetAuthorUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Body:                    request.GetBody(),
	}
}

func ListCommentsRequestToQuery(request *contentv1.ListCommentsRequest) usecase.ListCommentsQuery {
	return usecase.ListCommentsQuery{
		PostID:                  request.GetPostId(),
		ViewerUserID:            request.GetViewerUserId(),
		ViewerSubscriptionLevel: request.ViewerSubscriptionLevel,
		Limit:                   request.GetLimit(),
		Offset:                  request.GetOffset(),
	}
}

func NewListAuthorPostsResponse(posts []domain.PostSummary) *contentv1.ListAuthorPostsResponse {
	response := &contentv1.ListAuthorPostsResponse{
		Posts: make([]*contentv1.PostSummary, 0, len(posts)),
	}
	for _, post := range posts {
		response.Posts = append(response.Posts, postSummaryToProto(post))
	}

	return response
}

func NewSearchPostsResponse(posts []domain.PostSummary) *contentv1.SearchPostsResponse {
	response := &contentv1.SearchPostsResponse{
		Posts: make([]*contentv1.PostSummary, 0, len(posts)),
	}
	for _, post := range posts {
		response.Posts = append(response.Posts, postSummaryToProto(post))
	}

	return response
}

func NewPostResponse(post domain.Post) *contentv1.PostResponse {
	return &contentv1.PostResponse{
		Post: postToProto(post),
	}
}

func NewPostMediaResponse(media domain.PostMedia) *contentv1.PostMediaResponse {
	return &contentv1.PostMediaResponse{
		Media: &contentv1.PostMedia{
			FileUrl:     media.FileURL,
			Kind:        blockKindToProto(media.Kind),
			ContentType: media.ContentType,
			SizeBytes:   media.SizeBytes,
		},
	}
}

func NewPostLikeStateResponse(state domain.PostLikeState) *contentv1.PostLikeStateResponse {
	return &contentv1.PostLikeStateResponse{
		State: &contentv1.PostLikeState{
			PostId:     state.PostID,
			LikesCount: state.LikesCount,
			IsLiked:    state.IsLiked,
		},
	}
}

func NewCommentResponse(comment domain.Comment) *contentv1.CommentResponse {
	return &contentv1.CommentResponse{
		Comment: commentToProto(comment),
	}
}

func NewListCommentsResponse(comments []domain.Comment) *contentv1.ListCommentsResponse {
	response := &contentv1.ListCommentsResponse{
		Comments: make([]*contentv1.Comment, 0, len(comments)),
	}
	for _, comment := range comments {
		response.Comments = append(response.Comments, commentToProto(comment))
	}

	return response
}

func Empty() *emptypb.Empty {
	return &emptypb.Empty{}
}

func ErrorToStatus(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, usecase.ErrInvalidPostID),
		errors.Is(err, usecase.ErrInvalidUserID),
		errors.Is(err, usecase.ErrInvalidTitle),
		errors.Is(err, usecase.ErrInvalidRequiredSubscriptionLevel),
		errors.Is(err, usecase.ErrConflictingSubscriptionLevelUpdate),
		errors.Is(err, usecase.ErrBlocksRequired),
		errors.Is(err, usecase.ErrTooManyBlocks),
		errors.Is(err, usecase.ErrReplaceBlocksRequired),
		errors.Is(err, usecase.ErrInvalidLimit),
		errors.Is(err, usecase.ErrInvalidOffset),
		errors.Is(err, usecase.ErrInvalidSearchFilter),
		errors.Is(err, usecase.ErrInvalidCommentBody),
		errors.Is(err, usecase.ErrPostMediaFileNameRequired),
		errors.Is(err, usecase.ErrPostMediaContentTypeRequired),
		errors.Is(err, usecase.ErrPostMediaContentRequired),
		errors.Is(err, usecase.ErrPostMediaTooLarge),
		errors.Is(err, usecase.ErrPostMediaContentTypeUnsupported),
		errors.Is(err, domain.ErrInvalidBlockKind),
		errors.Is(err, domain.ErrInvalidBlockData):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPostNotFound), errors.Is(err, domain.ErrCommentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrPostForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, usecase.ErrPostMediaStorageUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func postBlockInputsFromProto(blocks []*contentv1.PostBlockInput) []usecase.PostBlockInput {
	result := make([]usecase.PostBlockInput, 0, len(blocks))
	for _, block := range blocks {
		result = append(result, usecase.PostBlockInput{
			Kind:        blockKindFromProto(block.GetKind()),
			TextContent: block.TextContent,
			FileURL:     block.FileUrl,
		})
	}

	return result
}

func postToProto(post domain.Post) *contentv1.Post {
	response := &contentv1.Post{
		PostId:        post.PostID,
		AuthorUserId:  post.AuthorUserID,
		Title:         post.Title,
		CreatedAt:     timestamppb.New(post.CreatedAt),
		UpdatedAt:     timestamppb.New(post.UpdatedAt),
		CanView:       post.CanView,
		LikesCount:    post.LikesCount,
		IsLiked:       post.IsLiked,
		CommentsCount: post.CommentsCount,
		Blocks:        make([]*contentv1.PostBlock, 0, len(post.Blocks)),
	}
	if post.RequiredSubscriptionLevel != nil {
		response.RequiredSubscriptionLevel = post.RequiredSubscriptionLevel
	}
	for _, block := range post.Blocks {
		response.Blocks = append(response.Blocks, postBlockToProto(block))
	}

	return response
}

func postSummaryToProto(post domain.PostSummary) *contentv1.PostSummary {
	response := &contentv1.PostSummary{
		PostId:        post.PostID,
		AuthorUserId:  post.AuthorUserID,
		Title:         post.Title,
		CreatedAt:     timestamppb.New(post.CreatedAt),
		CanView:       post.CanView,
		LikesCount:    post.LikesCount,
		IsLiked:       post.IsLiked,
		CommentsCount: post.CommentsCount,
	}
	if post.RequiredSubscriptionLevel != nil {
		response.RequiredSubscriptionLevel = post.RequiredSubscriptionLevel
	}

	return response
}

func postBlockToProto(block domain.PostBlock) *contentv1.PostBlock {
	response := &contentv1.PostBlock{
		PostBlockId: block.PostBlockID,
		Position:    block.Position,
		Kind:        blockKindToProto(block.Kind),
	}
	if block.TextContent != nil {
		response.TextContent = block.TextContent
	}
	if block.FileURL != nil {
		response.FileUrl = block.FileURL
	}

	return response
}

func commentToProto(comment domain.Comment) *contentv1.Comment {
	return &contentv1.Comment{
		CommentId:    comment.CommentID,
		PostId:       comment.PostID,
		AuthorUserId: comment.AuthorUserID,
		Body:         comment.Body,
		CreatedAt:    timestamppb.New(comment.CreatedAt),
		UpdatedAt:    timestamppb.New(comment.UpdatedAt),
	}
}

func blockKindFromProto(kind contentv1.ContentBlockKind) domain.BlockKind {
	switch kind {
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT:
		return domain.BlockKindText
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE:
		return domain.BlockKindImage
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO:
		return domain.BlockKindVideo
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT:
		return domain.BlockKindDocument
	default:
		return domain.BlockKind("")
	}
}

func blockKindsFromProto(kinds []contentv1.ContentBlockKind) []domain.BlockKind {
	result := make([]domain.BlockKind, 0, len(kinds))
	for _, kind := range kinds {
		result = append(result, blockKindFromProto(kind))
	}

	return result
}

func blockKindToProto(kind domain.BlockKind) contentv1.ContentBlockKind {
	switch kind {
	case domain.BlockKindText:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT
	case domain.BlockKindImage:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE
	case domain.BlockKindVideo:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO
	case domain.BlockKindDocument:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT
	default:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_UNSPECIFIED
	}
}
