package httpgateway_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
)

func TestDocsHandlerRendersSwaggerUIPage(t *testing.T) {
	handler := httpgateway.DocsHandler([]httpgateway.DocsSpec{
		{Name: "Auth API", URL: "/api/openapi/auth.swagger.json"},
		{Name: "Profile API", URL: "/api/openapi/profile.swagger.json"},
		{Name: "Content API", URL: "/api/openapi/content.swagger.json"},
	})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/docs/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", recorder.Code)
	}
	if contentType := recorder.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("unexpected content type: %s", contentType)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Fatalf("expected Swagger UI bundle in response body")
	}
	if !strings.Contains(body, "/api/openapi/auth.swagger.json") ||
		!strings.Contains(body, "/api/openapi/profile.swagger.json") ||
		!strings.Contains(body, "/api/openapi/content.swagger.json") {
		t.Fatalf("expected OpenAPI endpoints in response body: %s", body)
	}
}
