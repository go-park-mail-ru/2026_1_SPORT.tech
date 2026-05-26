package httpgateway

import (
	"encoding/json"
	"net/http"
	"strings"
)

var gatewayOpenAPITagAliases = map[string]string{
	"AuthService":         "Auth",
	"ProfileService":      "Profile",
	"PostService":         "Post",
	"TierService":         "Tier",
	"SubscriptionService": "Subscription",
	"SportService":        "Sport",
	"DonationService":     "Donation",
	"PaymentService":      "Payment",
	"StatisticsService":   "Statistics",
	"NotificationService": "Notification",
	"ChatService":         "Chat",
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
	rewritePostMediaUploadOperation(document)

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
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}

func rewriteAvatarUploadOperation(document map[string]any) {
	rewriteMultipartUploadOperation(document, "/v1/profiles/me/avatar", "avatar")
}

func rewritePostMediaUploadOperation(document map[string]any) {
	rewriteMultipartUploadOperation(document, "/v1/posts/media", "file")
}

func rewriteMultipartUploadOperation(document map[string]any, path string, fieldName string) {
	paths, ok := document["paths"].(map[string]any)
	if !ok {
		return
	}

	pathItem, ok := paths[path].(map[string]any)
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
			"name":     fieldName,
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
