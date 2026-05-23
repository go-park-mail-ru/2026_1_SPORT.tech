package domain

import "time"

type ChatMessage struct {
	MessageID      int64
	SenderUserID   int64
	ReceiverUserID int64
	Body           string
	IsRead         bool
	CreatedAt      time.Time
}

type ChatConversation struct {
	OtherUserID int64
	LastMessage ChatMessage
	UnreadCount int32
}
