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
	"google.golang.org/protobuf/encoding/protojson"
)

var gatewayOpenAPITagAliases = map[string]string{
	"AuthService":     "Auth",
	"ProfileService":  "Profile",
	"PostService":     "Post",
	"SportService":    "Sport",
	"DonationService": "Donation",
}

const gatewayOpenAPIBasePath = "/api"

func NewMux(
	ctx context.Context,
	authServer gatewayv1.AuthServiceServer,
	profileServer gatewayv1.ProfileServiceServer,
	postServer gatewayv1.PostServiceServer,
	sportServer gatewayv1.SportServiceServer,
	donationServer gatewayv1.DonationServiceServer,
) (http.Handler, error) {
	mux := newMux()

	if err := gatewayv1.RegisterAuthServiceHandlerServer(ctx, mux, authServer); err != nil {
		return nil, err
	}
	if err := gatewayv1.RegisterProfileServiceHandlerServer(ctx, mux, profileServer); err != nil {
		return nil, err
	}
	if err := gatewayv1.RegisterPostServiceHandlerServer(ctx, mux, postServer); err != nil {
		return nil, err
	}
	if err := gatewayv1.RegisterSportServiceHandlerServer(ctx, mux, sportServer); err != nil {
		return nil, err
	}
	if err := gatewayv1.RegisterDonationServiceHandlerServer(ctx, mux, donationServer); err != nil {
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
		runtime.WithMarshalerOption(
			runtime.MIMEWildcard,
			&runtime.HTTPBodyMarshaler{
				Marshaler: &runtime.JSONPb{
					MarshalOptions: protojson.MarshalOptions{UseProtoNames: true},
					UnmarshalOptions: protojson.UnmarshalOptions{
						DiscardUnknown: true,
					},
				},
			},
		),
		runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(outgoingHeaderMatcher),
		runtime.WithMetadata(incomingMetadata),
		runtime.WithForwardResponseOption(forwardResponseOption),
		runtime.WithErrorHandler(httpErrorHandler),
		runtime.WithRoutingErrorHandler(routingErrorHandler),
		runtime.WithMiddlewares(routePatternMiddleware(gatewayOpenAPIBasePath)),
	)
}

func routePatternMiddleware(prefix string) runtime.Middleware {
	return func(next runtime.HandlerFunc) runtime.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request, pathParams map[string]string) {
			pattern, ok := runtime.HTTPPathPattern(request.Context())
			if ok {
				if prefix != "" {
					pattern = prefix + pattern
				}

				if setter, ok := writer.(interface{ SetRoutePattern(string) }); ok {
					setter.SetRoutePattern(pattern)
				}
			}

			next(writer, request, pathParams)
		}
	}
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
	if metadataPairs.Get("x-session-token") == nil {
		if sessionCookie, err := request.Cookie("sid"); err == nil && strings.TrimSpace(sessionCookie.Value) != "" {
			metadataPairs.Set("x-session-token", sessionCookie.Value)
		}
	}

	return metadataPairs
}

func outgoingHeaderMatcher(key string) (string, bool) {
	return "", false
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
	addCSRFHeaders(document)
	rewriteAvatarUploadOperation(document)

	return json.Marshal(document)
}

func addCSRFHeaders(document map[string]any) {
	paths, ok := document["paths"].(map[string]any)
	if !ok {
		return
	}

	for path, rawPathItem := range paths {
		pathItem, ok := rawPathItem.(map[string]any)
		if !ok {
			continue
		}

		for method, rawOperation := range pathItem {
			operation, ok := rawOperation.(map[string]any)
			if !ok {
				continue
			}

			if !requiresOpenAPICSRFHeader(method, path) {
				continue
			}

			parameters, _ := operation["parameters"].([]any)
			parameters = append(parameters, csrfHeaderParameter())
			operation["parameters"] = parameters
		}
	}
}

func requiresOpenAPICSRFHeader(method string, path string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func rewriteAvatarUploadOperation(document map[string]any) {
	paths, ok := document["paths"].(map[string]any)
	if !ok {
		return
	}

	pathItem, ok := paths["/v1/profiles/me/avatar"].(map[string]any)
	if !ok {
		return
	}

	postOperation, ok := pathItem["post"].(map[string]any)
	if !ok {
		return
	}

	postOperation["consumes"] = []any{"multipart/form-data"}
	postOperation["parameters"] = []any{
		map[string]any{
			"name":     "avatar",
			"in":       "formData",
			"required": true,
			"type":     "file",
		},
		csrfHeaderParameter(),
	}
}

func csrfHeaderParameter() map[string]any {
	return map[string]any{
		"name":        csrfHeaderName,
		"in":          "header",
		"required":    true,
		"type":        "string",
		"description": "CSRF token from the csrf_token cookie",
	}
}
