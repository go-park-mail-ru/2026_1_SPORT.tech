PROTO_DIR := grpc/proto
PROTO_GEN_GO_DIR := grpc/gen/go
PROTO_GEN_OPENAPI_DIR := grpc/gen/openapiv2
PROTO_SERVICE_DIRS := $(PROTO_DIR)/auth $(PROTO_DIR)/profile $(PROTO_DIR)/content $(PROTO_DIR)/gateway
PROTO_FILES := $(shell find $(PROTO_SERVICE_DIRS) -name '*.proto' | sort)
COVER_PACKAGES := $(shell go list ./... | grep -E '/internal/(domain|usecase|adapters/mappers|infrastructure/httpgateway)$$' | grep -v '/grpc/gen/' | grep -v '/internal/mocks')
GO_BIN := $(HOME)/go/bin
COVERAGE_MIN ?= 60

.PHONY: generate
generate: proto easyjson mocks

.PHONY: mocks
mocks:
	PATH="$(GO_BIN):$$PATH" go generate ./services/auth/internal/mocks/... ./services/content/internal/mocks/... ./services/profile/internal/mocks/...

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

.PHONY: easyjson
easyjson:
	go generate ./...

.PHONY: test
test:
	go test ./...

.PHONY: fmt-check
fmt-check:
	@unformatted=$$(gofmt -l $$(find . -name '*.go' -not -path './.git/*')); \
	if [ -n "$$unformatted" ]; then \
		printf 'gofmt required for:\n%s\n' "$$unformatted"; \
		exit 1; \
	fi

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint: fmt-check vet

.PHONY: coverage
coverage:
	go test -covermode=atomic -coverprofile=coverage.tmp $(COVER_PACKAGES)
	grep -v -E '(/internal/mocks/|/grpc/gen/|_easyjson\.go|easyjson\.go|\.pb\.go|\.pb\.gw\.go|_grpc\.pb\.go)' coverage.tmp > coverage.out
	go tool cover -func=coverage.out | tail -n 1

.PHONY: coverage-total
coverage-total:
	@coverage_tmp=$$(mktemp); \
	coverage_out=$$(mktemp); \
	if ! go test -covermode=atomic -coverprofile=$$coverage_tmp $(COVER_PACKAGES) >/dev/null; then \
		rm -f $$coverage_tmp $$coverage_out; \
		exit 1; \
	fi; \
	grep -v -E '(/internal/mocks/|/grpc/gen/|_easyjson\.go|easyjson\.go|\.pb\.go|\.pb\.gw\.go|_grpc\.pb\.go)' $$coverage_tmp > $$coverage_out; \
	go tool cover -func=$$coverage_out | awk '/^total:/ { gsub("%", "", $$3); print $$3 }'; \
	rm -f $$coverage_tmp $$coverage_out

.PHONY: coverage-check
coverage-check:
	@coverage=$$(make --no-print-directory coverage-total); \
	printf 'coverage: %s%%\n' "$$coverage"; \
	awk -v coverage="$$coverage" -v min="$(COVERAGE_MIN)" 'BEGIN { if (coverage + 0 < min + 0) { printf "coverage %.1f%% is below required %.1f%%\n", coverage, min; exit 1 } }'

.PHONY: coverage-html
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

.PHONY: docker-build
docker-build:
	docker build -f services/auth/Dockerfile -t 2026_1_sporttech-auth-service .
	docker build -f services/profile/Dockerfile -t 2026_1_sporttech-profile-service .
	docker build -f services/content/Dockerfile -t 2026_1_sporttech-content-service .
	docker build -f services/api-gateway/Dockerfile -t 2026_1_sporttech-api-gateway .

.PHONY: ci
ci: lint test coverage-check

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
