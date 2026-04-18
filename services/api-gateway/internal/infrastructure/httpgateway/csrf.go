package httpgateway

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
)

var csrfExemptPaths = map[string]struct{}{
	"/api/v1/auth/login":            {},
	"/api/v1/auth/register/client":  {},
	"/api/v1/auth/register/trainer": {},
}

func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.HasPrefix(request.URL.Path, "/api/v1/") {
			next.ServeHTTP(writer, request)
			return
		}

		if shouldBootstrapCSRFCookie(request) {
			if csrfToken, err := newCSRFToken(); err == nil {
				setCSRFCookieOnResponse(writer, csrfToken)
			}
		}

		if !requiresCSRFProtection(request) {
			next.ServeHTTP(writer, request)
			return
		}

		sessionCookie, err := request.Cookie("sid")
		if err != nil || strings.TrimSpace(sessionCookie.Value) == "" {
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
	}

	_, exempt := csrfExemptPaths[request.URL.Path]
	return !exempt
}

func shouldBootstrapCSRFCookie(request *http.Request) bool {
	switch request.Method {
	case http.MethodGet, http.MethodHead:
	default:
		return false
	}

	sessionCookie, err := request.Cookie("sid")
	if err != nil || strings.TrimSpace(sessionCookie.Value) == "" {
		return false
	}

	csrfCookie, err := request.Cookie(csrfCookieName)
	return err != nil || strings.TrimSpace(csrfCookie.Value) == ""
}

func setCSRFCookieOnResponse(writer http.ResponseWriter, csrfToken string) {
	http.SetCookie(writer, &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	writer.Header().Set(csrfHeaderName, csrfToken)
}

func newCSRFToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
