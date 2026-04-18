package mappers

import (
	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
)

func RegisterRequestToAuth(request *gatewayv1.RegisterRequest) *authv1.RegisterRequest {
	return &authv1.RegisterRequest{
		Email:    request.GetEmail(),
		Username: request.GetUsername(),
		Password: request.GetPassword(),
		Role:     authv1.UserRole(request.GetRole()),
	}
}

func LoginRequestToAuth(request *gatewayv1.LoginRequest) *authv1.LoginRequest {
	return &authv1.LoginRequest{
		Email:    request.GetEmail(),
		Password: request.GetPassword(),
	}
}

func LogoutRequestToAuth(request *gatewayv1.LogoutRequest) *authv1.LogoutRequest {
	return &authv1.LogoutRequest{SessionToken: request.GetSessionToken()}
}

func ResolveSessionRequestToAuth(request *gatewayv1.ResolveSessionRequest) *authv1.GetSessionRequest {
	return &authv1.GetSessionRequest{SessionToken: request.GetSessionToken()}
}

func AuthSessionResponseFromAuth(response *authv1.AuthSessionResponse) *gatewayv1.AuthSessionResponse {
	if response == nil {
		return nil
	}

	return &gatewayv1.AuthSessionResponse{
		User:    authUserFromAuth(response.GetUser()),
		Session: sessionInfoFromAuth(response.GetSession()),
	}
}

func ResolveSessionResponseFromAuth(response *authv1.GetSessionResponse) *gatewayv1.ResolveSessionResponse {
	if response == nil {
		return nil
	}

	return &gatewayv1.ResolveSessionResponse{
		User:    authUserFromAuth(response.GetUser()),
		Session: sessionInfoFromAuth(response.GetSession()),
	}
}

func CreateProfileRequestToProfile(request *gatewayv1.CreateProfileRequest) *profilev1.CreateProfileRequest {
	return &profilev1.CreateProfileRequest{
		UserId:         request.GetUserId(),
		Username:       request.GetUsername(),
		FirstName:      request.GetFirstName(),
		LastName:       request.GetLastName(),
		Bio:            request.Bio,
		IsTrainer:      request.GetIsTrainer(),
		TrainerDetails: trainerDetailsToProfile(request.GetTrainerDetails()),
	}
}

func GetProfileRequestToProfile(request *gatewayv1.GetProfileRequest) *profilev1.GetProfileRequest {
	return &profilev1.GetProfileRequest{UserId: request.GetUserId()}
}

func UpdateProfileRequestToProfile(request *gatewayv1.UpdateProfileRequest) *profilev1.UpdateProfileRequest {
	return &profilev1.UpdateProfileRequest{
		UserId:         request.GetUserId(),
		Username:       request.Username,
		FirstName:      request.FirstName,
		LastName:       request.LastName,
		Bio:            request.Bio,
		TrainerDetails: trainerDetailsToProfile(request.GetTrainerDetails()),
	}
}

func SearchAuthorsRequestToProfile(request *gatewayv1.SearchAuthorsRequest) *profilev1.SearchAuthorsRequest {
	return &profilev1.SearchAuthorsRequest{
		Query:        request.GetQuery(),
		SportTypeIds: request.GetSportTypeIds(),
		Limit:        request.GetLimit(),
		Offset:       request.GetOffset(),
	}
}

func UploadAvatarRequestToProfile(request *gatewayv1.UploadAvatarRequest) *profilev1.UploadAvatarRequest {
	return &profilev1.UploadAvatarRequest{
		UserId:      request.GetUserId(),
		FileName:    request.GetFileName(),
		ContentType: request.GetContentType(),
		Content:     request.GetContent(),
	}
}

func DeleteAvatarRequestToProfile(request *gatewayv1.DeleteAvatarRequest) *profilev1.DeleteAvatarRequest {
	return &profilev1.DeleteAvatarRequest{UserId: request.GetUserId()}
}

func ProfileResponseFromProfile(response *profilev1.ProfileResponse) *gatewayv1.ProfileResponse {
	if response == nil {
		return nil
	}

	return &gatewayv1.ProfileResponse{Profile: profileFromProfile(response.GetProfile())}
}

func SearchAuthorsResponseFromProfile(response *profilev1.SearchAuthorsResponse) *gatewayv1.SearchAuthorsResponse {
	if response == nil {
		return nil
	}

	authors := make([]*gatewayv1.AuthorSummary, 0, len(response.GetAuthors()))
	for _, author := range response.GetAuthors() {
		authors = append(authors, authorSummaryFromProfile(author))
	}

	return &gatewayv1.SearchAuthorsResponse{Authors: authors}
}

func ListSportTypesResponseFromProfile(response *profilev1.ListSportTypesResponse) *gatewayv1.ListSportTypesResponse {
	if response == nil {
		return nil
	}

	sportTypes := make([]*gatewayv1.SportType, 0, len(response.GetSportTypes()))
	for _, sportType := range response.GetSportTypes() {
		sportTypes = append(sportTypes, &gatewayv1.SportType{
			SportTypeId: sportType.GetSportTypeId(),
			Name:        sportType.GetName(),
		})
	}

	return &gatewayv1.ListSportTypesResponse{SportTypes: sportTypes}
}

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
		PostId:                     request.GetPostId(),
		AuthorUserId:               request.GetAuthorUserId(),
		Title:                      request.Title,
		RequiredSubscriptionLevel:  request.RequiredSubscriptionLevel,
		ClearRequiredSubscriptionLevel: request.GetClearRequiredSubscriptionLevel(),
		Blocks:                     postBlockInputsToContent(request.GetBlocks()),
		ReplaceBlocks:              request.GetReplaceBlocks(),
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

func authUserFromAuth(user *authv1.AuthUser) *gatewayv1.AuthUser {
	if user == nil {
		return nil
	}

	return &gatewayv1.AuthUser{
		UserId:   user.GetUserId(),
		Email:    user.GetEmail(),
		Username: user.GetUsername(),
		Role:     gatewayv1.UserRole(user.GetRole()),
		Status:   gatewayv1.AccountStatus(user.GetStatus()),
	}
}

func sessionInfoFromAuth(session *authv1.SessionInfo) *gatewayv1.SessionInfo {
	if session == nil {
		return nil
	}

	return &gatewayv1.SessionInfo{
		SessionToken: session.GetSessionToken(),
		ExpiresAt:    session.GetExpiresAt(),
	}
}

func trainerDetailsToProfile(details *gatewayv1.TrainerDetails) *profilev1.TrainerDetails {
	if details == nil {
		return nil
	}

	sports := make([]*profilev1.TrainerSport, 0, len(details.GetSports()))
	for _, sport := range details.GetSports() {
		sports = append(sports, &profilev1.TrainerSport{
			SportTypeId:     sport.GetSportTypeId(),
			ExperienceYears: sport.GetExperienceYears(),
			SportsRank:      sport.SportsRank,
		})
	}

	return &profilev1.TrainerDetails{
		EducationDegree: details.EducationDegree,
		CareerSinceDate: details.CareerSinceDate,
		Sports:          sports,
	}
}

func trainerDetailsFromProfile(details *profilev1.TrainerDetails) *gatewayv1.TrainerDetails {
	if details == nil {
		return nil
	}

	sports := make([]*gatewayv1.TrainerSport, 0, len(details.GetSports()))
	for _, sport := range details.GetSports() {
		sports = append(sports, &gatewayv1.TrainerSport{
			SportTypeId:     sport.GetSportTypeId(),
			ExperienceYears: sport.GetExperienceYears(),
			SportsRank:      sport.SportsRank,
		})
	}

	return &gatewayv1.TrainerDetails{
		EducationDegree: details.EducationDegree,
		CareerSinceDate: details.CareerSinceDate,
		Sports:          sports,
	}
}

func profileFromProfile(profile *profilev1.Profile) *gatewayv1.Profile {
	if profile == nil {
		return nil
	}

	return &gatewayv1.Profile{
		UserId:         profile.GetUserId(),
		Username:       profile.GetUsername(),
		FirstName:      profile.GetFirstName(),
		LastName:       profile.GetLastName(),
		Bio:            profile.Bio,
		AvatarUrl:      profile.AvatarUrl,
		IsTrainer:      profile.GetIsTrainer(),
		CreatedAt:      profile.GetCreatedAt(),
		UpdatedAt:      profile.GetUpdatedAt(),
		TrainerDetails: trainerDetailsFromProfile(profile.GetTrainerDetails()),
	}
}

func authorSummaryFromProfile(author *profilev1.AuthorSummary) *gatewayv1.AuthorSummary {
	if author == nil {
		return nil
	}

	return &gatewayv1.AuthorSummary{
		UserId:         author.GetUserId(),
		Username:       author.GetUsername(),
		FirstName:      author.GetFirstName(),
		LastName:       author.GetLastName(),
		Bio:            author.Bio,
		AvatarUrl:      author.AvatarUrl,
		TrainerDetails: trainerDetailsFromProfile(author.GetTrainerDetails()),
	}
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
