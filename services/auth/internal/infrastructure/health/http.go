package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Pinger interface {
	PingContext(ctx context.Context) error
}

type response struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

func NewHandler(serviceName string, pinger Pinger) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()

		statusCode := http.StatusOK
		payload := response{
			Status:  "ok",
			Service: serviceName,
		}

		if err := pinger.PingContext(ctx); err != nil {
			statusCode = http.StatusServiceUnavailable
			payload.Status = "degraded"
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(statusCode)
		_ = json.NewEncoder(writer).Encode(payload)
	})
}
