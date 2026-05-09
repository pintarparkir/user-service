# ParkirPintar Makefile — orchestrates dev workflow
.PHONY: help up down logs seed proto test test-unit test-integration test-e2e \
        build smoke fmt vet lint mocks migrate clean

APP_NAME := parkirpintar
SERVICES := gateway reservation billing payment presence notification user

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Local stack ────────────────────────────────────────────────────────────────
up: ## Start full local stack (postgres, redis, rabbitmq, all services)
	podman compose -f deployments/docker-compose.yml up -d --build

down: ## Stop & remove local stack
	docker compose -f deployments/docker-compose.yml down -v

logs: ## Tail logs from all services
	docker compose -f deployments/docker-compose.yml logs -f

seed: ## Seed parking spots (150 cars + 250 motorcycles)
	docker compose -f deployments/docker-compose.yml exec -T postgres psql -U postgres -d parkirpintar -f /docker-entrypoint-initdb.d/02_seed.sql

smoke: ## Health check all services
	@for s in $(SERVICES); do \
	  echo "→ $$s"; \
	  curl -sf http://localhost:8080/healthz/$$s || echo "DOWN: $$s"; \
	done

# ── Code ───────────────────────────────────────────────────────────────────────
proto: ## Generate Go code from .proto files
	./scripts/gen_proto.sh

build: ## Compile all service binaries
	@for s in $(SERVICES) worker; do \
	  echo "→ building $$s"; \
	  go build -o bin/$$s ./cmd/$$s; \
	done

mocks: ## Regenerate mocks (mockgen)
	./scripts/gen_mocks.sh

migrate: ## Apply DB migrations
	./scripts/migrate.sh up

# ── Quality ────────────────────────────────────────────────────────────────────
fmt: ## Format Go code
	gofmt -s -w .

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint (requires install)
	golangci-lint run ./...

# ── Tests ──────────────────────────────────────────────────────────────────────
test: test-unit test-integration ## Run unit + integration

test-unit: ## Unit tests (no infra)
	go test -short -race -count=1 ./pkg/... ./internal/...

test-integration: ## Integration tests (requires `make up`)
	go test -race -count=1 -tags=integration ./test/integration/...

test-e2e: ## End-to-end tests (requires `make up && make seed`)
	go test -race -count=1 -tags=e2e -timeout=5m ./test/e2e/...

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out
