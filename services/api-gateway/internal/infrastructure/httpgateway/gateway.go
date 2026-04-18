package httpgateway

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

var gatewayOpenAPITagAliases = map[string]string{
	"AuthService":           "Auth Service",
	"ProfileService":        "Profile Service",
	"ContentService":        "Content Service",
	"GatewayAuthService":    "Auth Service",
	"GatewayProfileService": "Profile Service",
	"GatewayContentService": "Content Service",
}

const gatewayOpenAPIBasePath = "/api"

func NewMux(
	ctx context.Context,
	authServer gatewayv1.AuthServiceServer,
	profileServer gatewayv1.ProfileServiceServer,
	contentServer gatewayv1.ContentServiceServer,
) (http.Handler, error) {
	mux := newMux()

	if err := gatewayv1.RegisterAuthServiceHandlerServer(ctx, mux, authServer); err != nil {
		return nil, err
	}
	if err := gatewayv1.RegisterProfileServiceHandlerServer(ctx, mux, profileServer); err != nil {
		return nil, err
	}
	if err := gatewayv1.RegisterContentServiceHandlerServer(ctx, mux, contentServer); err != nil {
		return nil, err
	}

	return mux, nil
}

func OpenAPIHandler(filePath string, tagAliases map[string]string, basePath string) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := os.Stat(filePath); err != nil {
			http.NotFound(writer, request)
			return
		}

		writer.Header().Set("Content-Type", "application/json")

		if len(tagAliases) == 0 {
			http.ServeFile(writer, request, filePath)
			return
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(writer, "failed to read openapi spec", http.StatusInternalServerError)
			return
		}

		normalized, err := rewriteOpenAPISpec(data, tagAliases, basePath)
		if err != nil {
			http.Error(writer, "failed to normalize openapi spec", http.StatusInternalServerError)
			return
		}

		_, _ = writer.Write(normalized)
	})
}

func GatewayOpenAPIHandler(filePath string) http.Handler {
	return OpenAPIHandler(filePath, gatewayOpenAPITagAliases, gatewayOpenAPIBasePath)
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

func rewriteOpenAPISpec(data []byte, tagAliases map[string]string, basePath string) ([]byte, error) {
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, err
	}

	if tags, ok := document["tags"].([]any); ok {
		for _, rawTag := range tags {
			tag, ok := rawTag.(map[string]any)
			if !ok {
				continue
			}

			name, ok := tag["name"].(string)
			if !ok {
				continue
			}
			if alias, ok := tagAliases[name]; ok {
				tag["name"] = alias
			}
		}
	}

	paths, ok := document["paths"].(map[string]any)
	if ok {
		for _, rawPathItem := range paths {
			pathItem, ok := rawPathItem.(map[string]any)
			if !ok {
				continue
			}

			for _, rawOperation := range pathItem {
				operation, ok := rawOperation.(map[string]any)
				if !ok {
					continue
				}

				tags, ok := operation["tags"].([]any)
				if !ok {
					continue
				}

				for index, rawTag := range tags {
					name, ok := rawTag.(string)
					if !ok {
						continue
					}
					if alias, ok := tagAliases[name]; ok {
						tags[index] = alias
					}
				}
			}
		}
	}

	if basePath != "" {
		document["basePath"] = basePath
	}

	return json.Marshal(document)
}
