package httpgateway

import (
	"net/http"
	"strings"
)

const (
	corsAllowMethods = "GET, HEAD, OPTIONS, POST, PATCH, DELETE"
	corsExposeHeader = "X-CSRF-Token"
	corsDefaultAllow = "Content-Type, X-CSRF-Token, X-Request-Id"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := strings.TrimSpace(request.Header.Get("Origin"))
		if origin == "" {
			next.ServeHTTP(writer, request)
			return
		}

		headers := writer.Header()
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Set("Access-Control-Allow-Methods", corsAllowMethods)
		headers.Set("Access-Control-Expose-Headers", corsExposeHeader)

		allowHeaders := strings.TrimSpace(request.Header.Get("Access-Control-Request-Headers"))
		if allowHeaders == "" {
			allowHeaders = corsDefaultAllow
		}
		headers.Set("Access-Control-Allow-Headers", allowHeaders)

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
