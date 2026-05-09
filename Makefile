.PHONY: help run build test test-unit test-integration lint vet fmt mocks proto \
        migrate-up migrate-down docker-build docker-run down clean

PROJECT_ENV ?= local

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Run / build ────────────────────────────────────────────────────────────────
run: ## Run user-service against local infra (postgres, redis, otel)
	PROJECT_ENV=$(PROJECT_ENV) go run ./cmd/user

build: ## Compile binary into bin/user
	CGO_ENABLED=0 go build -o bin/user ./cmd/user

# ── Tests ──────────────────────────────────────────────────────────────────────
test: ## Run all tests
	go test ./...

test-unit: ## Unit tests only (no infra)
	go test -short -race -count=1 ./pkg/... ./internal/...

test-integration: ## Integration tests (requires postgres on localhost:5432)
	go test -race -count=1 -tags=integration ./test/integration/...

# ── Quality ────────────────────────────────────────────────────────────────────
fmt: ## Format Go code
	gofmt -s -w .

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint (requires install)
	golangci-lint run ./...

# ── Code generation ────────────────────────────────────────────────────────────
proto: ## Regenerate api/proto/user/v1/{user.pb.go, user_grpc.pb.go}
	@which protoc >/dev/null || (echo "protoc not installed (brew install protobuf)" && exit 1)
	@which protoc-gen-go >/dev/null || (echo "protoc-gen-go not installed (go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)" && exit 1)
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/proto/user/v1/user.proto

mocks: ## Regenerate gomock test doubles under mock/
	./scripts/gen_mocks.sh

# ── DB migrations ──────────────────────────────────────────────────────────────
migrate-up: ## Apply DB migrations
	./scripts/migrate.sh up

migrate-down: ## Roll back one migration
	./scripts/migrate.sh down 1

# ── Container / docker compose ────────────────────────────────────────────────
docker-build: ## Build the service container image
	docker build -t user-service:local .

docker-run: ## Bring up the service via deployments/docker-compose.yml (expects ../infra up)
	docker compose -f deployments/docker-compose.yml up --build

down: ## Tear down the service's docker compose stack
	docker compose -f deployments/docker-compose.yml down

# ── Housekeeping ───────────────────────────────────────────────────────────────
clean: ## Remove build artefacts
	rm -rf bin/ coverage.out
