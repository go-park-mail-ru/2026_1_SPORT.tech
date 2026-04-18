package httpgateway_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
)

func TestCSRFMiddlewareRejectsUnsafeRequestWithoutToken(t *testing.T) {
	handler := httpgateway.CSRFMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/posts", nil)
	request.AddCookie(&http.Cookie{Name: "sid", Value: "session-token"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

func TestCSRFMiddlewareAllowsUnsafeRequestWithMatchingToken(t *testing.T) {
	handler := httpgateway.CSRFMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodPost, "/api/v1/posts", nil)
	request.AddCookie(&http.Cookie{Name: "sid", Value: "session-token"})
	request.AddCookie(&http.Cookie{Name: "csrf_token", Value: "csrf-value"})
	request.Header.Set("X-CSRF-Token", "csrf-value")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

func TestCSRFMiddlewareBootstrapsTokenOnSafeRequest(t *testing.T) {
	handler := httpgateway.CSRFMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "sid", Value: "session-token"})

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if csrfHeader := recorder.Header().Get("X-CSRF-Token"); strings.TrimSpace(csrfHeader) == "" {
		t.Fatalf("expected X-CSRF-Token header")
	}
	if setCookie := recorder.Header().Values("Set-Cookie"); len(setCookie) == 0 || !strings.Contains(strings.Join(setCookie, ";"), "csrf_token=") {
		t.Fatalf("expected csrf cookie, got %q", setCookie)
	}
}
