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

func (handler *Handler) requestContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := newRequestID()

		baseLogger := handler.logger
		if baseLogger == nil {
			baseLogger = slog.Default()
		}

		requestLogger := baseLogger.With("request_id", requestID)

		ctx := requestctx.WithRequestID(request.Context(), requestID)
		ctx = requestctx.WithLogger(ctx, requestLogger)

		writer.Header().Set(requestIDHeader, requestID)

		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func newRequestID() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	return hex.EncodeToString(buffer)
}
