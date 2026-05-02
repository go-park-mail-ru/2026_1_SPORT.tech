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
			{"name":"AuthService"},
			{"name":"ProfileService"},
			{"name":"PostService"},
			{"name":"SportService"},
			{"name":"DonationService"}
		],
		"paths":{
			"/v1/auth/login":{"post":{"tags":["AuthService"]}},
			"/v1/profiles/{user_id}":{"get":{"tags":["ProfileService"]}},
			"/v1/posts/{post_id}":{"get":{"tags":["PostService"]}},
			"/v1/sport-types":{"get":{"tags":["SportService"]}},
			"/v1/profiles/{user_id}/donations":{"post":{"tags":["DonationService"]}}
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

	if !gotTags["Auth"] || !gotTags["Profile"] || !gotTags["Post"] || !gotTags["Sport"] || !gotTags["Donation"] {
		t.Fatalf("unexpected tags: %#v", gotTags)
	}
}

func TestGatewayOpenAPIHandlerAddsCSRFOnlyToUnsafeMethods(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	specPath := filepath.Join(dir, "gateway.swagger.json")
	spec := `{
		"swagger":"2.0",
		"tags":[{"name":"AuthService"},{"name":"PostService"}],
		"paths":{
			"/v1/auth/csrf":{"get":{"tags":["AuthService"]}},
			"/v1/auth/login":{"post":{"tags":["AuthService"]}},
			"/v1/posts":{"post":{"tags":["PostService"]}}
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

	paths := payload["paths"].(map[string]any)
	csrfGet := paths["/v1/auth/csrf"].(map[string]any)["get"].(map[string]any)
	if _, ok := csrfGet["parameters"]; ok {
		t.Fatalf("did not expect csrf header on GET /v1/auth/csrf: %#v", csrfGet["parameters"])
	}

	loginPost := paths["/v1/auth/login"].(map[string]any)["post"].(map[string]any)
	if !hasHeaderParameter(loginPost["parameters"], "X-CSRF-Token") {
		t.Fatalf("expected csrf header on POST /v1/auth/login: %#v", loginPost["parameters"])
	}

	postsPost := paths["/v1/posts"].(map[string]any)["post"].(map[string]any)
	if !hasHeaderParameter(postsPost["parameters"], "X-CSRF-Token") {
		t.Fatalf("expected csrf header on POST /v1/posts: %#v", postsPost["parameters"])
	}
}

func TestGatewayOpenAPIHandlerRewritesPostMediaUpload(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	specPath := filepath.Join(dir, "gateway.swagger.json")
	spec := `{
		"swagger":"2.0",
		"tags":[{"name":"PostService"}],
		"paths":{
			"/v1/posts/media":{"post":{"tags":["PostService"],"parameters":[]}}
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

	paths := payload["paths"].(map[string]any)
	postOperation := paths["/v1/posts/media"].(map[string]any)["post"].(map[string]any)
	if !hasFormFileParameter(postOperation["parameters"], "file") {
		t.Fatalf("expected multipart file parameter: %#v", postOperation["parameters"])
	}
	if !hasHeaderParameter(postOperation["parameters"], "X-CSRF-Token") {
		t.Fatalf("expected csrf header parameter: %#v", postOperation["parameters"])
	}
}

func hasHeaderParameter(raw any, name string) bool {
	parameters, ok := raw.([]any)
	if !ok {
		return false
	}

	for _, rawParameter := range parameters {
		parameter, ok := rawParameter.(map[string]any)
		if !ok {
			continue
		}
		if parameter["in"] == "header" && parameter["name"] == name {
			return true
		}
	}

	return false
}

func hasFormFileParameter(raw any, name string) bool {
	parameters, ok := raw.([]any)
	if !ok {
		return false
	}

	for _, rawParameter := range parameters {
		parameter, ok := rawParameter.(map[string]any)
		if !ok {
			continue
		}
		if parameter["in"] == "formData" && parameter["name"] == name && parameter["type"] == "file" {
			return true
		}
	}

	return false
}
