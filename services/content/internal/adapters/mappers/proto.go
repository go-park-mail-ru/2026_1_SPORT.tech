package mappers

import (
	"errors"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/domain"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func Empty() *emptypb.Empty {
	return &emptypb.Empty{}
}

func ErrorToStatus(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, usecase.ErrInvalidPostID),
		errors.Is(err, usecase.ErrInvalidUserID),
		errors.Is(err, usecase.ErrInvalidTitle),
		errors.Is(err, usecase.ErrInvalidRequiredSubscriptionLevel),
		errors.Is(err, usecase.ErrConflictingSubscriptionLevelUpdate),
		errors.Is(err, usecase.ErrInvalidSportTypeID),
		errors.Is(err, usecase.ErrConflictingSportTypeUpdate),
		errors.Is(err, usecase.ErrBlocksRequired),
		errors.Is(err, usecase.ErrTooManyBlocks),
		errors.Is(err, usecase.ErrReplaceBlocksRequired),
		errors.Is(err, usecase.ErrInvalidLimit),
		errors.Is(err, usecase.ErrInvalidOffset),
		errors.Is(err, usecase.ErrInvalidSearchFilter),
		errors.Is(err, usecase.ErrInvalidCommentBody),
		errors.Is(err, usecase.ErrPostMediaFileNameRequired),
		errors.Is(err, usecase.ErrPostMediaContentTypeRequired),
		errors.Is(err, usecase.ErrPostMediaContentRequired),
		errors.Is(err, usecase.ErrPostMediaTooLarge),
		errors.Is(err, usecase.ErrPostMediaContentTypeUnsupported),
		errors.Is(err, usecase.ErrInvalidSubscriptionTierID),
		errors.Is(err, usecase.ErrInvalidSubscriptionTierName),
		errors.Is(err, usecase.ErrInvalidSubscriptionTierPrice),
		errors.Is(err, usecase.ErrInvalidSubscriptionTierDescription),
		errors.Is(err, usecase.ErrConflictingTierDescriptionUpdate),
		errors.Is(err, usecase.ErrInvalidSubscriptionID),
		errors.Is(err, usecase.ErrInvalidSubscriptionTarget),
		errors.Is(err, usecase.ErrInvalidDonationAmount),
		errors.Is(err, usecase.ErrInvalidDonationCurrency),
		errors.Is(err, usecase.ErrInvalidDonationMessage),
		errors.Is(err, usecase.ErrInvalidDonationTarget),
		errors.Is(err, usecase.ErrInvalidPaymentID),
		errors.Is(err, usecase.ErrInvalidPaymentConfirmationToken),
		errors.Is(err, usecase.ErrInvalidNotificationID),
		errors.Is(err, usecase.ErrInvalidNotificationType),
		errors.Is(err, usecase.ErrInvalidNotificationTitle),
		errors.Is(err, usecase.ErrInvalidNotificationBody),
		errors.Is(err, domain.ErrInvalidBlockKind),
		errors.Is(err, domain.ErrInvalidBlockData):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrPostNotFound),
		errors.Is(err, domain.ErrCommentNotFound),
		errors.Is(err, domain.ErrSubscriptionTierNotFound),
		errors.Is(err, domain.ErrSubscriptionNotFound),
		errors.Is(err, domain.ErrDonationNotFound),
		errors.Is(err, domain.ErrPaymentNotFound),
		errors.Is(err, domain.ErrNotificationNotFound),
		errors.Is(err, domain.ErrChatMessageNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrPostForbidden),
		errors.Is(err, domain.ErrPaymentForbidden),
		errors.Is(err, domain.ErrPaymentTokenMismatch),
		errors.Is(err, domain.ErrChatAccessForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrSubscriptionTierInUse),
		errors.Is(err, domain.ErrPaymentAlreadyConfirmed),
		errors.Is(err, usecase.ErrPaymentNotSucceeded),
		errors.Is(err, usecase.ErrSubscriptionPaymentRequired):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, usecase.ErrPostMediaStorageUnavailable),
		errors.Is(err, usecase.ErrPaymentProviderUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}

func blockKindFromProto(kind contentv1.ContentBlockKind) domain.BlockKind {
	switch kind {
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT:
		return domain.BlockKindText
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE:
		return domain.BlockKindImage
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO:
		return domain.BlockKindVideo
	case contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT:
		return domain.BlockKindDocument
	default:
		return domain.BlockKind("")
	}
}

func blockKindsFromProto(kinds []contentv1.ContentBlockKind) []domain.BlockKind {
	result := make([]domain.BlockKind, 0, len(kinds))
	for _, kind := range kinds {
		result = append(result, blockKindFromProto(kind))
	}

	return result
}

func blockKindToProto(kind domain.BlockKind) contentv1.ContentBlockKind {
	switch kind {
	case domain.BlockKindText:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT
	case domain.BlockKindImage:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE
	case domain.BlockKindVideo:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_VIDEO
	case domain.BlockKindDocument:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_DOCUMENT
	default:
		return contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_UNSPECIFIED
	}
}

func postBlockInputsFromProto(blocks []*contentv1.PostBlockInput) []usecase.PostBlockInput {
	result := make([]usecase.PostBlockInput, 0, len(blocks))
	for _, block := range blocks {
		result = append(result, usecase.PostBlockInput{
			Kind:        blockKindFromProto(block.GetKind()),
			TextContent: block.TextContent,
			FileURL:     block.FileUrl,
		})
	}

	return result
}
