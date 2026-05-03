package domain

import "errors"

var (
	ErrPostNotFound             = errors.New("post not found")
	ErrPostForbidden            = errors.New("post forbidden")
	ErrCommentNotFound          = errors.New("comment not found")
	ErrSubscriptionTierNotFound = errors.New("subscription tier not found")
	ErrSubscriptionTierInUse    = errors.New("subscription tier is used by posts")
	ErrInvalidBlockKind         = errors.New("invalid block kind")
	ErrInvalidBlockData         = errors.New("invalid block data")
)
