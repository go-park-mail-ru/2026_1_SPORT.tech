package httpgateway

import (
	"net/http"
	"strings"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
)

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.HasPrefix(request.URL.Path, "/api/v1/") {
			next.ServeHTTP(writer, request)
			return
		}

		if !requiresCSRFProtection(request) {
			next.ServeHTTP(writer, request)
			return
		}

		csrfCookie, err := request.Cookie(csrfCookieName)
		if err != nil || strings.TrimSpace(csrfCookie.Value) == "" {
			writePublicError(writer, http.StatusForbidden, "forbidden", "csrf token is required")
			return
		}

		csrfHeader := strings.TrimSpace(request.Header.Get(csrfHeaderName))
		if csrfHeader == "" || csrfHeader != csrfCookie.Value {
			writePublicError(writer, http.StatusForbidden, "forbidden", "invalid csrf token")
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func requiresCSRFProtection(request *http.Request) bool {
	switch request.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
