PROTO_DIR := grpc/proto
PROTO_GEN_GO_DIR := grpc/gen/go
PROTO_GEN_OPENAPI_DIR := grpc/gen/openapiv2
PROTO_FILES := $(shell find $(PROTO_DIR) -name '*.proto')
GO_BIN := $(HOME)/go/bin

AUTH_CONFIG_PATH ?= services/auth/configs/service.yml
AUTH_DB_URL ?= postgres://postgres:postgres@localhost:5432/sporttech_auth?sslmode=disable
BIN_DIR ?= bin

.PHONY: tools generate proto auth-build auth-run auth-test auth-test-integration auth-migrate-up auth-migrate-down

tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.28.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.28.0

generate: proto

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

auth-build:
	mkdir -p $(BIN_DIR)
	GOSUMDB=off GOPROXY=off go build -o ./$(BIN_DIR)/auth-service ./services/auth/cmd/service

auth-run:
	GOSUMDB=off GOPROXY=off AUTH_CONFIG_PATH=$(AUTH_CONFIG_PATH) go run ./services/auth/cmd/service

auth-test:
	GOSUMDB=off GOPROXY=off go test ./services/auth/...

auth-test-integration:
	GOSUMDB=off GOPROXY=off go test -tags integration ./services/auth/internal/adapters/repository/postgres/...

auth-migrate-up:
	migrate -path services/auth/migrations -database "$(AUTH_DB_URL)" up

auth-migrate-down:
	migrate -path services/auth/migrations -database "$(AUTH_DB_URL)" down 1
