package mappers

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
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
		SportTypeIDs:                 request.GetSportTypeIds(),
		BlockKinds:                   blockKindsFromProto(request.GetBlockKinds()),
		MinRequiredSubscriptionLevel: request.MinRequiredSubscriptionLevel,
		MaxRequiredSubscriptionLevel: request.MaxRequiredSubscriptionLevel,
		OnlyAvailable:                request.GetOnlyAvailable(),
		ViewerUserID:                 request.GetViewerUserId(),
		ViewerSubscriptionLevel:      request.ViewerSubscriptionLevel,
		Limit:                        request.GetLimit(),
		Offset:                       request.GetOffset(),
		Sort:                         request.GetSort(),
	}
}

func CreatePostRequestToCommand(request *contentv1.CreatePostRequest) usecase.CreatePostCommand {
	return usecase.CreatePostCommand{
		AuthorUserID:              request.GetAuthorUserId(),
		Title:                     request.GetTitle(),
		RequiredSubscriptionLevel: request.RequiredSubscriptionLevel,
		SportTypeID:               request.SportTypeId,
		SportTypeIDs:              request.GetSportTypeIds(),
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
		SportTypeID:                    request.SportTypeId,
		ClearSportTypeID:               request.GetClearSportTypeId(),
		SportTypeIDs:                   request.GetSportTypeIds(),
		ClearSportTypeIDs:              request.GetClearSportTypeIds(),
		Blocks:                         postBlockInputsFromProto(request.GetBlocks()),
		ReplaceBlocks:                  request.GetReplaceBlocks(),
		IsPinned:                       request.IsPinned,
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
		IsPinned:      post.IsPinned,
		Blocks:        make([]*contentv1.PostBlock, 0, len(post.Blocks)),
	}
	if post.RequiredSubscriptionLevel != nil {
		response.RequiredSubscriptionLevel = post.RequiredSubscriptionLevel
	}
	if post.SportTypeID != nil {
		response.SportTypeId = post.SportTypeID
	}
	response.SportTypeIds = post.SportTypeIDs
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
		IsPinned:      post.IsPinned,
	}
	if post.RequiredSubscriptionLevel != nil {
		response.RequiredSubscriptionLevel = post.RequiredSubscriptionLevel
	}
	if post.SportTypeID != nil {
		response.SportTypeId = post.SportTypeID
	}
	response.SportTypeIds = post.SportTypeIDs

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
