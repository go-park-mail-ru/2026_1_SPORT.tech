PROTO_DIR := grpc/proto
PROTO_GEN_GO_DIR := grpc/gen/go
PROTO_GEN_OPENAPI_DIR := grpc/gen/openapiv2
PROTO_SERVICE_DIRS := $(PROTO_DIR)/auth $(PROTO_DIR)/profile $(PROTO_DIR)/content $(PROTO_DIR)/gateway
PROTO_FILES := $(shell find $(PROTO_SERVICE_DIRS) -name '*.proto' | sort)
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
		--openapiv2_out=allow_merge=false:$(PROTO_GEN_OPENAPI_DIR) \
		$(PROTO_FILES)
	rm -rf $(PROTO_GEN_OPENAPI_DIR)/google $(PROTO_GEN_OPENAPI_DIR)/protoc-gen-openapiv2

.PHONY: auth-build
auth-build:
	mkdir -p bin
	GOSUMDB=off GOPROXY=off go build -o ./bin/auth-service ./services/auth/cmd/service

.PHONY: auth-run
auth-run:
	GOSUMDB=off GOPROXY=off AUTH_CONFIG_PATH=services/auth/configs/service.yml go run ./services/auth/cmd/service

.PHONY: auth-test
auth-test:
	GOSUMDB=off GOPROXY=off go test ./services/auth/...

.PHONY: auth-test-integration
auth-test-integration:
	GOSUMDB=off GOPROXY=off go test -tags integration ./services/auth/internal/adapters/repository/postgres/...

.PHONY: profile-build
profile-build:
	mkdir -p bin
	GOSUMDB=off GOPROXY=off go build -o ./bin/profile-service ./services/profile/cmd/service

.PHONY: profile-run
profile-run:
	GOSUMDB=off GOPROXY=off PROFILE_CONFIG_PATH=services/profile/configs/service.yml go run ./services/profile/cmd/service

.PHONY: profile-test
profile-test:
	GOSUMDB=off GOPROXY=off go test ./services/profile/...

.PHONY: profile-test-integration
profile-test-integration:
	GOSUMDB=off GOPROXY=off go test -tags integration ./services/profile/internal/adapters/repository/postgres/...

.PHONY: content-build
content-build:
	mkdir -p bin
	GOSUMDB=off GOPROXY=off go build -o ./bin/content-service ./services/content/cmd/service

.PHONY: content-run
content-run:
	GOSUMDB=off GOPROXY=off CONTENT_CONFIG_PATH=services/content/configs/service.yml go run ./services/content/cmd/service

.PHONY: content-test
content-test:
	GOSUMDB=off GOPROXY=off go test ./services/content/...

.PHONY: content-test-integration
content-test-integration:
	GOSUMDB=off GOPROXY=off go test -tags integration ./services/content/internal/adapters/repository/postgres/...

.PHONY: api-gateway-build
api-gateway-build:
	mkdir -p bin
	GOSUMDB=off GOPROXY=off go build -o ./bin/api-gateway ./services/api-gateway/cmd/service

.PHONY: api-gateway-run
api-gateway-run:
	GOSUMDB=off GOPROXY=off API_GATEWAY_CONFIG_PATH=services/api-gateway/configs/service.yml go run ./services/api-gateway/cmd/service

.PHONY: api-gateway-test
api-gateway-test:
	GOSUMDB=off GOPROXY=off go test ./services/api-gateway/...

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
