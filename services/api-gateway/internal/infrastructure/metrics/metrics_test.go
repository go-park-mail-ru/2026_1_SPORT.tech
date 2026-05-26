package metrics

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestNormalizedHTTPPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "static route",
			path: "/api/v1/posts:search",
			want: "/api/v1/posts:search",
		},
		{
			name: "numeric id",
			path: "/api/v1/posts/123",
			want: "/api/v1/posts/{id}",
		},
		{
			name: "nested numeric id",
			path: "/api/v1/trainers/1001/tiers",
			want: "/api/v1/trainers/{id}/tiers",
		},
		{
			name: "hex id",
			path: "/api/v1/files/7c33a8a8882f53ad0f199e6e40f677c8",
			want: "/api/v1/files/{id}",
		},
		{
			name: "plain slug",
			path: "/api/v1/sport-types",
			want: "/api/v1/sport-types",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := normalizedHTTPPath(test.path); got != test.want {
				t.Fatalf("normalizedHTTPPath(%q) = %q, want %q", test.path, got, test.want)
			}
		})
	}
}

func TestSanitizeHTTPMetricPathRejectsInvalidUTF8(t *testing.T) {
	if got := sanitizeHTTPMetricPath("/api/\xff"); got != invalidHTTPMetricPath {
		t.Fatalf("sanitizeHTTPMetricPath() = %q, want %q", got, invalidHTTPMetricPath)
	}
}

func TestHTTPMiddlewarePreservesFlush(t *testing.T) {
	metricsSet := New("test-service")

	var flushErr error
	handler := metricsSet.HTTPMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		flushErr = http.NewResponseController(writer).Flush()
	}))

	request := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: "/api/v1/chat/messages/1/stream"}}
	handler.ServeHTTP(httptest.NewRecorder(), request)

	if flushErr != nil {
		t.Fatalf("ResponseController.Flush() through metrics middleware failed: %v (statusRecorder must implement Unwrap so SSE streams can flush)", flushErr)
	}
}

func TestHTTPMiddlewareHandlesInvalidUTF8Path(t *testing.T) {
	metricsSet := New("test-service")
	handler := metricsSet.HTTPMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	}))

	request := &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Path: "/\xff;{curl,http://example.oast.online};"},
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("HTTPMiddleware panicked on invalid UTF-8 path: %v", recovered)
		}
	}()

	handler.ServeHTTP(httptest.NewRecorder(), request)
}
