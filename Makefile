# Developer entry points. Run `make help` for an overview.
SHELL := /bin/bash
export PATH := $(PATH):/usr/local/go/bin:/usr/local/node/bin:$(HOME)/go/bin

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BIN     := bin/trisse-home
WEB     := web

# Native packages (ADR 0015). PKG_VERSION is a Debian version derived from
# the latest release tag, see scripts/package-version.sh.
NFPM        ?= nfpm
PKG_ARCHS   ?= amd64 arm64
PKG_VERSION ?= $(shell sh scripts/package-version.sh)

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk -F':.*?## ' '{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

.PHONY: deps
deps: ## Install Go modules and frontend dependencies
	go mod download
	cd $(WEB) && npm ci --no-audit --no-fund

.PHONY: generate
generate: ## Generate TypeScript DTO types (tygo) and the backend message catalogue
	go run ./scripts/tygoconfig -o .tygo.generated.yaml
	tygo generate --config .tygo.generated.yaml
	rm -f .tygo.generated.yaml
	go run ./scripts/servermessages
	cd $(WEB) && npx prettier --log-level warn --write src/shared/api-client/generated >/dev/null || true

.PHONY: lint
lint: lint-go lint-web ## Run all linters

.PHONY: lint-go
lint-go: ## Lint Go code (golangci-lint incl. depguard)
	golangci-lint run ./...

.PHONY: lint-web
lint-web: ## Lint and type-check the frontend
	cd $(WEB) && npm run lint && npm run check

.PHONY: test
test: test-go test-web ## Run all unit and integration tests

.PHONY: test-go
test-go: ## Run Go tests with the race detector
	go test -race -count=1 ./...

.PHONY: test-web
test-web: ## Run frontend unit tests (Vitest)
	cd $(WEB) && npm test

.PHONY: cover
cover: ## Go test coverage report
	go test -count=1 -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out | tail -1

.PHONY: build-web
build-web: ## Build the frontend into internal/platform/webui/dist
	cd $(WEB) && npm run build

.PHONY: build
build: build-web ## Build the single binary with the embedded frontend
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" -o $(BIN) ./cmd/server

.PHONY: package
package: build-web ## Build .deb packages for PKG_ARCHS into dist/ (needs nfpm)
	mkdir -p dist
	set -e; for arch in $(PKG_ARCHS); do \
		CGO_ENABLED=0 GOOS=linux GOARCH=$$arch go build -trimpath -ldflags="-s -w -X main.version=$(VERSION)" \
			-o bin/linux-$$arch/trisse-home ./cmd/server; \
		GOARCH=$$arch PKG_VERSION=$(PKG_VERSION) $(NFPM) package --config packaging/nfpm.yaml --packager deb --target dist/; \
	done
	cd dist && sha256sum *.deb > trisse-home_$(PKG_VERSION)_packages.sha256

.PHONY: run
run: build ## Run locally with data in ./data (needs SECRET_KEY or .env)
	set -a; [ -f .env ] && source .env; set +a; \
	DATABASE_PATH=$${DATABASE_PATH_LOCAL:-data/trisse-home.db} LOG_FORMAT=text ./$(BIN)

.PHONY: smoke
smoke: build ## Run the Playwright smoke test against a fresh instance
	cd $(WEB) && npm run test:e2e

.PHONY: check-generated
check-generated: generate ## Fail if generated files are out of date
	@test -z "$$(git status --porcelain -- $(WEB)/src/shared/api-client/generated)" || \
		{ git status --short -- $(WEB)/src/shared/api-client/generated; echo "generated files are out of date: run make generate"; exit 1; }

.PHONY: ci
ci: lint test check-generated build ## Everything CI runs

.PHONY: docker-build
docker-build: ## Build the container image
	sudo docker compose build --build-arg VERSION=$(VERSION)

.PHONY: up
up: ## Build and (re)start the stack
	sudo docker compose build --build-arg VERSION=$(VERSION) && sudo docker compose up -d

.PHONY: logs
logs: ## Follow the application logs
	sudo docker compose logs -f trisse-home

.PHONY: clean
clean: ## Remove build output
	rm -rf bin dist coverage.out internal/platform/webui/dist/assets internal/platform/webui/dist/index.html

new-module: ## Scaffold a new module: make new-module NAME=recipes [TITLE="Recipes"]
	go run ./scripts/newmodule -name "$(NAME)" -title "$(TITLE)"
	$(MAKE) generate
	cd web && npx prettier --log-level warn --write src
