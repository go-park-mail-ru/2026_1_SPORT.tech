package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/infrastructure/requestctx"
)

const requestIDHeader = "X-Request-ID"

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

func (handler *Handler) requestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		requestID := newRequestID()

		baseLogger := handler.logger
		if baseLogger == nil {
			baseLogger = slog.Default()
		}

		requestLogger := baseLogger.With("request_id", requestID)

		ctx := requestctx.WithRequestID(request.Context(), requestID)
		ctx = requestctx.WithLogger(ctx, requestLogger)

		responseWriter := &statusResponseWriter{
			ResponseWriter: writer,
			statusCode:     http.StatusOK,
		}

		responseWriter.Header().Set(requestIDHeader, requestID)

		next.ServeHTTP(responseWriter, request.WithContext(ctx))

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

func newRequestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	return hex.EncodeToString(buffer)
}
