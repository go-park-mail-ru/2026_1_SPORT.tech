package httpgateway_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
)

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	handler := httpgateway.CORSMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Fatalf("next handler must not be called for preflight")
	}))

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/me", nil)
	request.Header.Set("Origin", "http://212.233.98.238")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "X-CSRF-Token, Content-Type")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if allowOrigin := recorder.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "http://212.233.98.238" {
		t.Fatalf("unexpected allow origin: %q", allowOrigin)
	}
	if allowCredentials := recorder.Header().Get("Access-Control-Allow-Credentials"); allowCredentials != "true" {
		t.Fatalf("unexpected allow credentials: %q", allowCredentials)
	}
	if allowHeaders := recorder.Header().Get("Access-Control-Allow-Headers"); allowHeaders != "X-CSRF-Token, Content-Type" {
		t.Fatalf("unexpected allow headers: %q", allowHeaders)
	}
}

func TestCORSMiddlewareAddsHeadersToRegularRequest(t *testing.T) {
	handler := httpgateway.CORSMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	request.Header.Set("Origin", "http://212.233.98.238")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if allowOrigin := recorder.Header().Get("Access-Control-Allow-Origin"); allowOrigin != "http://212.233.98.238" {
		t.Fatalf("unexpected allow origin: %q", allowOrigin)
	}
	if exposeHeaders := recorder.Header().Get("Access-Control-Expose-Headers"); exposeHeaders != "X-CSRF-Token" {
		t.Fatalf("unexpected expose headers: %q", exposeHeaders)
	}
}
