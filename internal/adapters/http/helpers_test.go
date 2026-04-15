package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/requestctx"
)

func TestWriteHelpers(t *testing.T) {
	recorder := httptest.NewRecorder()
	writeValidationError(recorder, []validationErrorField{{Field: "username", Message: "bad"}})

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var response errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if response.Error.Code != "validation_error" || len(response.Error.Fields) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}

	recorder = httptest.NewRecorder()
	writeNoContent(recorder)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected no content status: %d", recorder.Code)
	}
}

func TestRequestMiddlewareSetsRequestID(t *testing.T) {
	handler := &Handler{logger: slog.Default()}

	next := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID, ok := requestctx.RequestIDFromContext(request.Context())
		if !ok || requestID == "" {
			t.Fatal("expected request id in context")
		}
		writer.WriteHeader(http.StatusCreated)
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.requestMiddleware(next).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if recorder.Header().Get(requestIDHeader) == "" {
		t.Fatal("expected X-Request-ID header")
	}
}

func TestNormalizePublicURL(t *testing.T) {
	handler := &Handler{storagePublicBaseURL: "http://example.com:8000/avatars"}

	rawURL := "http://localhost:8000/avatars/users/9/file.jpg"
	normalized := handler.normalizePublicURL(&rawURL)
	if normalized == nil || *normalized != "http://example.com:8000/avatars/users/9/file.jpg" {
		t.Fatalf("unexpected normalized url: %v", normalized)
	}

	remoteURL := "http://cdn.example.com/avatars/users/9/file.jpg"
	if normalized := handler.normalizePublicURL(&remoteURL); normalized == nil || *normalized != remoteURL {
		t.Fatalf("unexpected remote url normalization: %v", normalized)
	}
}

func TestCORSHelpers(t *testing.T) {
	if !isAllowedOrigin("http://example.com:3000", "example.com:8080") {
		t.Fatal("expected same hostname origin to be allowed")
	}
	if isAllowedOrigin("http://other.com:3000", "example.com:8080") {
		t.Fatal("expected different hostname origin to be rejected")
	}
	if requestHostname("example.com:8080") != "example.com" {
		t.Fatalf("unexpected hostname: %q", requestHostname("example.com:8080"))
	}
}

func TestCORSMiddlewareOptions(t *testing.T) {
	handler := &Handler{}
	request := httptest.NewRequest(http.MethodOptions, "/auth/me", nil)
	request.Header.Set("Origin", "http://example.com:3000")
	request.Host = "example.com:8080"
	recorder := httptest.NewRecorder()

	handler.corsMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Fatal("next handler should not be called for preflight")
	})).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatal("expected CORS methods header")
	}
}
