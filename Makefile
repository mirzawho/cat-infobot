# InfoBot — common development and release tasks.
#
#   make help      list targets
#   make run       start the bot locally (reads .env)
#   make check     gofmt + vet + tests
#   make docker    build the image
#   make release   tag and push; GitHub Actions publishes the image

BINARY     := bin/infobot
IMAGE      ?= ghcr.io/mirzawho/cat-infobot
TAG        ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
DATA_DIR   ?= $(HOME)/.local/share/infobot
GOCACHE    ?= $(or $(TMPDIR),/tmp)/go-build

GO         ?= go
DOCKER     ?= docker
COMPOSE    ?= docker compose
GOFMT      := $(GO) fmt
GOFLAGS_ENV := GOCACHE=$(GOCACHE)

.DEFAULT_GOAL := help
.PHONY: help build run check fmt vet test test-race tidy clean docker docker-run compose-up compose-down \
        compose-logs compose-config k8s-dry-run release install-tools

help: ## Show this help
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build bin/infobot
	@mkdir -p $(dir $(BINARY))
	$(GOFLAGS_ENV) $(GO) build -trimpath -o $(BINARY) .
	@echo "built ./$(BINARY)"

run: ## Run the bot from source (loads .env into the environment)
	@test -f .env || (echo "missing .env: run 'cp .env.example .env' and fill it in" >&2; exit 1)
	@set -a && . ./.env && set +a && \
		DATA_DIR="$${DATA_DIR:-$(DATA_DIR)}" $(GOFLAGS_ENV) $(GO) run .

check: fmt vet test ## gofmt, vet and tests

fmt: ## Format the code (fails if files change when FORMAT_CHECK=1)
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		if [ "$(FORMAT_CHECK)" = "1" ]; then echo "not gofmt-ed:"; echo "$$files"; exit 1; fi; \
		echo "gofmt: $$files"; $(GOFMT) ./...; \
	else \
		echo "gofmt: ok"; \
	fi

vet: ## Run go vet
	$(GOFLAGS_ENV) $(GO) vet ./...

test: ## Run tests
	$(GOFLAGS_ENV) $(GO) test -count=1 ./...

test-race: ## Run tests with the race detector
	$(GOFLAGS_ENV) $(GO) test -race -count=1 ./...

tidy: ## Tidy go.mod / go.sum
	$(GO) mod tidy
	$(GO) mod verify

clean: ## Remove build output
	rm -rf bin/

docker: ## Build the Docker image as $(IMAGE):$(TAG)
	$(DOCKER) build -t $(IMAGE):$(TAG) $(if $(filter dev HEAD,$(TAG)),,-t $(IMAGE):latest) .

docker-run: ## Run the image with .env (no ports, data volume)
	$(DOCKER) run --rm -d --name infobot --restart unless-stopped \
		--read-only --tmpfs /tmp --cap-drop ALL --security-opt no-new-privileges \
		--env-file .env -v infobot-data:/data $(IMAGE):$(TAG)

compose-up: ## Build and start with Docker Compose
	$(COMPOSE) up -d --build
	$(COMPOSE) ps

compose-down: ## Stop Docker Compose services
	$(COMPOSE) down

compose-logs: ## Follow Docker Compose logs
	$(COMPOSE) logs -f --tail=100 infobot

compose-config: ## Validate docker-compose.yml
	$(COMPOSE) config --quiet && echo "compose config: ok"

k8s-dry-run: ## Validate k8s.yaml against the cluster (requires kubectl)
	kubectl apply --dry-run=server -f k8s.yaml

release: ## Tag HEAD and push it (VERSION must be set); CI publishes the image
	@test -n "$(VERSION)" || (echo "usage: make release VERSION=v0.1.0" >&2; exit 1)
	@git diff --quiet && git diff --cached --quiet || (echo "working tree is dirty; commit first" >&2; exit 1)
	git tag -a "$(VERSION)" -m "InfoBot $(VERSION)"
	git push origin "$(VERSION)"
	@echo "pushed $(VERSION); GitHub Actions publishes $(IMAGE):$$(git rev-parse HEAD)"
