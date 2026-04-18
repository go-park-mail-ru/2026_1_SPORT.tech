package httpgateway

import (
	"context"
	"net/http"
	"os"
	"strings"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

func NewMux(ctx context.Context, server gatewayv1.GatewayServiceServer) (http.Handler, error) {
	mux := newMux()

	if err := gatewayv1.RegisterGatewayServiceHandlerServer(ctx, mux, server); err != nil {
		return nil, err
	}

	return mux, nil
}

func OpenAPIHandler(filePath string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := os.Stat(filePath); err != nil {
			http.NotFound(writer, request)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		http.ServeFile(writer, request, filePath)
	})
}

func newMux() *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher),
		runtime.WithMetadata(incomingMetadata),
	)
}

func incomingHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case "authorization", "x-request-id", "x-session-token", "x-user-id", "x-subscription-level":
		return strings.ToLower(key), true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func incomingMetadata(ctx context.Context, request *http.Request) metadata.MD {
	headers := []string{"authorization", "x-request-id", "x-session-token", "x-user-id", "x-subscription-level"}
	metadataPairs := metadata.MD{}
	for _, header := range headers {
		if value := request.Header.Get(header); value != "" {
			metadataPairs.Set(header, value)
		}
	}

	return metadataPairs
}
