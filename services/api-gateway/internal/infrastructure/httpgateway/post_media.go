package httpgateway

import (
	"context"
	"io"
	"net/http"
	"strings"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	postMediaUploadRoutePattern = "/api/v1/posts/media"
	maxPostMediaUploadBodySize  = 12 * 1024 * 1024
)

type PostMediaUploader interface {
	UploadPostMedia(context.Context, *gatewayv1.UploadPostMediaRequest) (*gatewayv1.PostMediaUploadResponse, error)
}

func MultipartPostMediaHandler(uploader PostMediaUploader, fallback http.Handler) http.Handler {
	marshalOptions := protojson.MarshalOptions{UseProtoNames: true}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			fallback.ServeHTTP(writer, request)
			return
		}

		if setter, ok := writer.(interface{ SetRoutePattern(string) }); ok {
			setter.SetRoutePattern(postMediaUploadRoutePattern)
		}
		request.Body = http.MaxBytesReader(writer, request.Body, maxPostMediaUploadBodySize)

		file, fileHeader, err := request.FormFile("file")
		if err != nil {
			if strings.Contains(err.Error(), "request body too large") {
				writePublicError(writer, http.StatusRequestEntityTooLarge, "payload_too_large", "file is too large")
				return
			}
			writePublicError(writer, http.StatusBadRequest, "bad_request", "file is required")
			return
		}
		defer file.Close()

		content, err := io.ReadAll(file)
		if err != nil {
			writePublicError(writer, http.StatusBadRequest, "bad_request", "failed to read file")
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")
		if contentType == "" || contentType == "application/octet-stream" {
			contentType = http.DetectContentType(content)
		}

		ctx := metadata.NewIncomingContext(request.Context(), incomingMetadata(request.Context(), request))
		response, err := uploader.UploadPostMedia(ctx, &gatewayv1.UploadPostMediaRequest{
			File:        content,
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
		writer.WriteHeader(http.StatusCreated)
		_, _ = writer.Write(payload)
	})
}
