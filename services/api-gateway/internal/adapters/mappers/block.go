package mappers

import (
	"strings"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
)

func blockKindToContent(kind string) contentv1.ContentBlockKind {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "text":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT
	case "image":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE
	case "video":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO
	case "document":
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT
	default:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_UNSPECIFIED
	}
}

func blockKindsToContent(kinds []string) []contentv1.ContentBlockKind {
	result := make([]contentv1.ContentBlockKind, 0, len(kinds))
	for _, kind := range kinds {
		result = append(result, blockKindToContent(kind))
	}

	return result
}

func blockKindFromContent(kind contentv1.ContentBlockKind) string {
	switch kind {
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT:
		return "text"
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE:
		return "image"
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO:
		return "video"
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT:
		return "document"
	default:
		return ""
	}
}

func postBlockInputsToContent(blocks []*gatewayv1.PostBlockInput) []*contentv1.PostBlockInput {
	result := make([]*contentv1.PostBlockInput, 0, len(blocks))
	for _, block := range blocks {
		if block == nil {
			continue
		}

		result = append(result, &contentv1.PostBlockInput{
			Kind:        blockKindToContent(block.GetKind()),
			TextContent: trimOptionalString(block.TextContent),
			FileUrl:     trimOptionalString(block.FileUrl),
		})
	}

	return result
}

func postBlocksFromContent(blocks []*contentv1.PostBlock) ([]*gatewayv1.PostBlock, error) {
	result := make([]*gatewayv1.PostBlock, 0, len(blocks))
	for _, block := range blocks {
		postBlockID, err := int64ToInt32("content.post_block.post_block_id", block.GetPostBlockId())
		if err != nil {
			return nil, err
		}

		result = append(result, &gatewayv1.PostBlock{
			PostBlockId: postBlockID,
			Position:    block.GetPosition(),
			Kind:        blockKindFromContent(block.GetKind()),
			TextContent: block.TextContent,
			FileUrl:     block.FileUrl,
		})
	}

	return result, nil
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
