package handler

import (
	"net"
	nethttp "net/http"
	"net/url"
	"strings"
)

var allowAllOrigins = true

func (handler *Handler) corsMiddleware(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(writer nethttp.ResponseWriter, request *nethttp.Request) {
		origin := request.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(writer, request)
			return
		}

		if !allowAllOrigins && !isAllowedOrigin(origin, request.Host) {
			if request.Method == nethttp.MethodOptions {
				writer.WriteHeader(nethttp.StatusForbidden)
				return
			}

			next.ServeHTTP(writer, request)
			return
		}

		headers := writer.Header()
		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		headers.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if request.Method == nethttp.MethodOptions {
			writer.WriteHeader(nethttp.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func isAllowedOrigin(origin string, requestHost string) bool {
	originURL, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if originURL.Scheme != "http" && originURL.Scheme != "https" {
		return false
	}

	if originURL.Hostname() == "" {
		return false
	}

	return strings.EqualFold(originURL.Hostname(), requestHostname(requestHost))
}

func requestHostname(host string) string {
	hostname, _, err := net.SplitHostPort(host)
	if err == nil {
		return hostname
	}

	return host
}
