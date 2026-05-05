package httpgateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
)

type stubPostMediaUploader struct {
	uploadFunc func(ctx context.Context, request *gatewayv1.UploadPostMediaRequest) (*gatewayv1.PostMediaUploadResponse, error)
}

func (stub stubPostMediaUploader) UploadPostMedia(ctx context.Context, request *gatewayv1.UploadPostMediaRequest) (*gatewayv1.PostMediaUploadResponse, error) {
	return stub.uploadFunc(ctx, request)
}

func TestMultipartPostMediaHandlerUploadsFile(t *testing.T) {
	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)
	fileWriter, err := multipartWriter.CreateFormFile("file", "run.png")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	content := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if _, err := fileWriter.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := multipartWriter.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	handler := httpgateway.MultipartPostMediaHandler(
		stubPostMediaUploader{
			uploadFunc: func(ctx context.Context, request *gatewayv1.UploadPostMediaRequest) (*gatewayv1.PostMediaUploadResponse, error) {
				if request.GetFileName() != "run.png" ||
					request.GetContentType() != "image/png" ||
					!bytes.Equal(request.GetFile(), content) {
					t.Fatalf("unexpected upload request: %+v", request)
				}

				return &gatewayv1.PostMediaUploadResponse{
					FileUrl:     "http://storage/post-media/posts/7/run.png",
					Kind:        "image",
					ContentType: "image/png",
					SizeBytes:   int32(len(content)),
				}, nil
			},
		},
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			t.Fatal("did not expect fallback handler")
		}),
	)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/posts/media", &body)
	request.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	var response struct {
		FileURL string `json:"file_url"`
		Kind    string `json:"kind"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.FileURL != "http://storage/post-media/posts/7/run.png" || response.Kind != "image" {
		t.Fatalf("unexpected response: %+v", response)
	}
}
