PROTO_DIR := grpc/proto
PROTO_GEN_GO_DIR := grpc/gen/go
PROTO_GEN_OPENAPI_DIR := grpc/gen/openapiv2
PROTO_SERVICE_DIRS := $(PROTO_DIR)/auth $(PROTO_DIR)/profile $(PROTO_DIR)/content $(PROTO_DIR)/gateway
PROTO_FILES := $(shell find $(PROTO_SERVICE_DIRS) -name '*.proto' | sort)
COVER_PACKAGES := $(shell go list ./... | grep -v '/grpc/gen/' | grep -v '/internal/mocks')
GO_BIN := $(HOME)/go/bin

.PHONY: generate
generate: proto

.PHONY: proto
proto:
	mkdir -p $(PROTO_GEN_GO_DIR) $(PROTO_GEN_OPENAPI_DIR)
	PATH="$(GO_BIN):$$PATH" protoc \
		-I $(PROTO_DIR) \
		-I /usr/include \
		--go_out=paths=source_relative:$(PROTO_GEN_GO_DIR) \
		--go-grpc_out=paths=source_relative,require_unimplemented_servers=false:$(PROTO_GEN_GO_DIR) \
		--grpc-gateway_out=paths=source_relative:$(PROTO_GEN_GO_DIR) \
		--openapiv2_out=allow_merge=false,json_names_for_fields=false:$(PROTO_GEN_OPENAPI_DIR) \
		$(PROTO_FILES)
	rm -rf $(PROTO_GEN_OPENAPI_DIR)/google $(PROTO_GEN_OPENAPI_DIR)/protoc-gen-openapiv2

.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -covermode=atomic -coverprofile=coverage.tmp $(COVER_PACKAGES)
	grep -v -E '(/internal/mocks/|/grpc/gen/|_easyjson\.go|easyjson\.go|\.pb\.go|\.pb\.gw\.go|_grpc\.pb\.go)' coverage.tmp > coverage.out
	go tool cover -func=coverage.out | tail -n 1

.PHONY: coverage-html
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

.PHONY: compose-up
compose-up:
	docker compose up --build -d

.PHONY: compose-down
compose-down:
	docker compose down

.PHONY: compose-logs
compose-logs:
	docker compose logs -f

.PHONY: compose-ps
compose-ps:
	docker compose ps
