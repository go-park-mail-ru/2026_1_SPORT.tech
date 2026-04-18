package httpgateway

import (
	"context"
	"io"
	"net/http"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

const avatarUploadRoutePattern = "/api/profiles/me/avatar"

type AvatarUploader interface {
	UploadMyAvatar(context.Context, *gatewayv1.UploadMyAvatarRequest) (*gatewayv1.AvatarUploadResponse, error)
}

func MultipartAvatarHandler(uploader AvatarUploader, fallback http.Handler) http.Handler {
	marshalOptions := protojson.MarshalOptions{UseProtoNames: true}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			fallback.ServeHTTP(writer, request)
			return
		}

		if setter, ok := writer.(interface{ SetRoutePattern(string) }); ok {
			setter.SetRoutePattern(avatarUploadRoutePattern)
		}

		file, fileHeader, err := request.FormFile("avatar")
		if err != nil {
			writePublicError(writer, http.StatusBadRequest, "bad_request", "avatar file is required")
			return
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			writePublicError(writer, http.StatusBadRequest, "bad_request", "failed to read avatar file")
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(content)
		}

		ctx := metadata.NewIncomingContext(request.Context(), incomingMetadata(request.Context(), request))
		response, err := uploader.UploadMyAvatar(ctx, &gatewayv1.UploadMyAvatarRequest{
			Avatar:      content,
			FileName:    fileHeader.Filename,
			ContentType: contentType,
		})
		if err != nil {
			writeGRPCError(writer, err)
			return
		}

		payload, err := marshalOptions.Marshal(response)
		if err != nil {
			writePublicError(writer, http.StatusInternalServerError, "internal_error", "failed to encode response")
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write(payload)
	})
}
