package usecase

import "errors"

var (
	ErrInvalidPostID                      = errors.New("invalid post id")
	ErrInvalidUserID                      = errors.New("invalid user id")
	ErrInvalidTitle                       = errors.New("invalid post title")
	ErrInvalidRequiredSubscriptionLevel   = errors.New("invalid required subscription level")
	ErrConflictingSubscriptionLevelUpdate = errors.New("required subscription level and clear flag cannot be used together")
	ErrBlocksRequired                     = errors.New("post blocks are required")
	ErrTooManyBlocks                      = errors.New("too many post blocks")
	ErrReplaceBlocksRequired              = errors.New("replace_blocks must be true when blocks are provided")
	ErrInvalidLimit                       = errors.New("invalid limit")
	ErrInvalidOffset                      = errors.New("invalid offset")
	ErrInvalidCommentBody                 = errors.New("invalid comment body")
)
