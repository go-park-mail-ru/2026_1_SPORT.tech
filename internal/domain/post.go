package domain

import "time"

type PostListItem struct {
	PostID    int64
	TrainerID int64
	MinTierID *int64
	Title     string
	CreatedAt time.Time
	CanView   bool
}

type PostAttachment struct {
	PostAttachmentID int64
	Kind             string
	FileURL          string
}

type PostLikeStatus struct {
	PostID     int64
	LikesCount int64
	IsLiked    bool
}

type Post struct {
	PostID      int64
	TrainerID   int64
	MinTierID   *int64
	Title       string
	TextContent string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CanView     bool
	Attachments []PostAttachment
}
