package domain

import "time"

type BlockKind string

const (
	BlockKindText     BlockKind = "text"
	BlockKindImage    BlockKind = "image"
	BlockKindVideo    BlockKind = "video"
	BlockKindDocument BlockKind = "document"
)

type PostBlock struct {
	PostBlockID int64
	Position    int32
	Kind        BlockKind
	TextContent *string
	FileURL     *string
}

type Post struct {
	PostID                    int64
	AuthorUserID              int64
	Title                     string
	RequiredSubscriptionLevel *int32
	SportTypeID               *int64
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	CanView                   bool
	LikesCount                int64
	IsLiked                   bool
	CommentsCount             int64
	Blocks                    []PostBlock
}

type PostSummary struct {
	PostID                    int64
	AuthorUserID              int64
	Title                     string
	RequiredSubscriptionLevel *int32
	SportTypeID               *int64
	CreatedAt                 time.Time
	CanView                   bool
	LikesCount                int64
	IsLiked                   bool
	CommentsCount             int64
}

type PostLikeState struct {
	PostID     int64
	LikesCount int64
	IsLiked    bool
}

type Comment struct {
	CommentID    int64
	PostID       int64
	AuthorUserID int64
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (kind BlockKind) IsValid() bool {
	switch kind {
	case BlockKindText, BlockKindImage, BlockKindVideo, BlockKindDocument:
		return true
	default:
		return false
	}
}

func CanViewPost(requiredLevel *int32, authorUserID int64, viewerUserID int64, viewerSubscriptionLevel *int32) bool {
	if requiredLevel == nil {
		return true
	}
	if authorUserID == viewerUserID && authorUserID != 0 {
		return true
	}
	if viewerSubscriptionLevel == nil {
		return false
	}

	return *viewerSubscriptionLevel >= *requiredLevel
}
