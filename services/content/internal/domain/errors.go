package domain

import "errors"

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrPostForbidden    = errors.New("post forbidden")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrInvalidBlockKind = errors.New("invalid block kind")
	ErrInvalidBlockData = errors.New("invalid block data")
)
