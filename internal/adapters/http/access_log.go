package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/requestctx"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (writer *statusResponseWriter) WriteHeader(statusCode int) {
	writer.statusCode = statusCode
	writer.ResponseWriter.WriteHeader(statusCode)
}

func (writer *statusResponseWriter) Write(body []byte) (int, error) {
	if writer.statusCode == 0 {
		writer.statusCode = http.StatusOK
	}

	return writer.ResponseWriter.Write(body)
}

func (handler *Handler) accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()

		responseWriter := &statusResponseWriter{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(responseWriter, request)

		requestLogger, ok := requestctx.LoggerFromContext(request.Context())
		if !ok || requestLogger == nil {
			requestLogger = handler.logger
		}
		if requestLogger == nil {
			requestLogger = slog.Default()
		}

		requestLogger.Info(
			"http request",
			"method", request.Method,
			"path", request.URL.Path,
			"status", responseWriter.statusCode,
			"duration_ms", time.Since(startedAt).Milliseconds(),
			"remote_addr", request.RemoteAddr,
		)
	})
}
