package httpgateway_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
)

func TestGatewayOpenAPIHandlerRewritesTags(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	specPath := filepath.Join(dir, "gateway.swagger.json")
	spec := `{
		"swagger":"2.0",
		"tags":[
			{"name":"GatewayAuthService"},
			{"name":"GatewayProfileService"},
			{"name":"GatewayContentService"}
		],
		"paths":{
			"/v1/auth/login":{"post":{"tags":["GatewayAuthService"]}},
			"/v1/profiles/{user_id}":{"get":{"tags":["GatewayProfileService"]}},
			"/v1/posts/{post_id}":{"get":{"tags":["GatewayContentService"]}}
		}
	}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/openapi/gateway.swagger.json", nil)

	httpgateway.GatewayOpenAPIHandler(specPath).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["basePath"] != "/api" {
		t.Fatalf("unexpected basePath: %#v", payload["basePath"])
	}

	tags := payload["tags"].([]any)
	gotTags := map[string]bool{}
	for _, rawTag := range tags {
		tag := rawTag.(map[string]any)
		gotTags[tag["name"].(string)] = true
	}

	if !gotTags["Auth Service"] || !gotTags["Profile Service"] || !gotTags["Content Service"] {
		t.Fatalf("unexpected tags: %#v", gotTags)
	}
}
