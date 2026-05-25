package usecase

import "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"

const (
	PostSortRecent  = "recent"
	PostSortPopular = "popular"
)

type PostBlockInput struct {
	Kind        domain.BlockKind
	TextContent *string
	FileURL     *string
}

type ListAuthorPostsQuery struct {
	AuthorUserID            int64
	ViewerUserID            int64
	ViewerSubscriptionLevel *int32
	Limit                   int32
	Offset                  int32
}

type SearchPostsQuery struct {
	Query                        string
	AuthorUserIDs                []int64
	SportTypeIDs                 []int64
	BlockKinds                   []domain.BlockKind
	MinRequiredSubscriptionLevel *int32
	MaxRequiredSubscriptionLevel *int32
	OnlyAvailable                bool
	ViewerUserID                 int64
	ViewerSubscriptionLevel      *int32
	Limit                        int32
	Offset                       int32
	Sort                         string
}

type GetPostQuery struct {
	PostID                  int64
	ViewerUserID            int64
	ViewerSubscriptionLevel *int32
}

type ListSubscriptionTiersQuery struct {
	TrainerUserID int64
}

type ListMySubscriptionsQuery struct {
	ClientUserID int64
}

type ListTrainerSubscribersQuery struct {
	TrainerUserID int64
	Limit         int32
	Offset        int32
}

type GetBalanceQuery struct {
	TrainerUserID int64
	Currency      string
}

type GetTrainerStatisticsQuery struct {
	TrainerUserID int64
	Currency      string
}

type ListCommentsQuery struct {
	PostID                  int64
	ViewerUserID            int64
	ViewerSubscriptionLevel *int32
	Limit                   int32
	Offset                  int32
}

type ListPostLikesQuery struct {
	PostID                  int64
	ViewerUserID            int64
	ViewerSubscriptionLevel *int32
	Limit                   int32
	Offset                  int32
}

type ListNotificationsQuery struct {
	UserID int64
	Limit  int32
	Offset int32
}

type ListReceivedDonationsQuery struct {
	TrainerUserID int64
	Limit         int32
	Offset        int32
}

type ListChatMessagesQuery struct {
	UserID      int64
	OtherUserID int64
	Limit       int32
	Offset      int32
}

type ListChatConversationsQuery struct {
	UserID int64
}

type ListMeetingAvailabilityRulesQuery struct {
	TrainerUserID int64
}

type ListMyMeetingSlotsQuery struct {
	TrainerUserID int64
}

type ListTrainerMeetingAvailabilityQuery struct {
	TrainerUserID int64
	ViewerUserID  int64
	From          string
	To            string
}

type ListMeetingsQuery struct {
	UserID int64
}
