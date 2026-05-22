package domain

import "errors"

var (
	ErrPostNotFound             = errors.New("post not found")
	ErrPostForbidden            = errors.New("post forbidden")
	ErrCommentNotFound          = errors.New("comment not found")
	ErrSubscriptionTierNotFound = errors.New("subscription tier not found")
	ErrSubscriptionTierInUse    = errors.New("subscription tier is used by posts")
	ErrSubscriptionNotFound     = errors.New("subscription not found")
	ErrDonationNotFound         = errors.New("donation not found")
	ErrPaymentNotFound          = errors.New("payment not found")
	ErrPaymentAlreadyConfirmed  = errors.New("payment already confirmed")
	ErrPaymentForbidden         = errors.New("payment forbidden")
	ErrPaymentTokenMismatch     = errors.New("payment confirmation token mismatch")
	ErrNotificationNotFound     = errors.New("notification not found")
	ErrInvalidBlockKind         = errors.New("invalid block kind")
	ErrInvalidBlockData         = errors.New("invalid block data")
)
