package httpgateway

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type DownstreamEndpoints struct {
	AuthGRPCEndpoint    string
	ProfileGRPCEndpoint string
	ContentGRPCEndpoint string
}

func NewMux(ctx context.Context, endpoints DownstreamEndpoints) (http.Handler, error) {
	mux := newMux()
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := authv1.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, endpoints.AuthGRPCEndpoint, dialOptions); err != nil {
		return nil, fmt.Errorf("register auth handlers: %w", err)
	}
	if err := profilev1.RegisterProfileServiceHandlerFromEndpoint(ctx, mux, endpoints.ProfileGRPCEndpoint, dialOptions); err != nil {
		return nil, fmt.Errorf("register profile handlers: %w", err)
	}
	if err := contentv1.RegisterContentServiceHandlerFromEndpoint(ctx, mux, endpoints.ContentGRPCEndpoint, dialOptions); err != nil {
		return nil, fmt.Errorf("register content handlers: %w", err)
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
