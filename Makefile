DOCKER ?= docker
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Image config
IMAGE_REPO ?= quay.io/stackdome/stackdome
VERSION ?= $(shell git rev-parse --short HEAD)
IMAGE_TAG ?= $(IMAGE_REPO):$(VERSION)

generate:
	rm -rf pkg/api/openapi
	$(DOCKER) run -v ${PWD}:/local:rw openapitools/openapi-generator-cli:v6.0.1 generate -i /local/config/openapi/stackdome_api.yaml -g go -o /local/pkg/api/openapi
	gofmt -w pkg/api/openapi
	rm pkg/api/openapi/go.mod
	rm pkg/api/openapi/go.sum
.PHONY: generate

frontend:
	pnpm --prefix frontend install --frozen-lockfile
	pnpm --prefix frontend exec vite build
	touch pkg/web/dist/.gitkeep
.PHONY: frontend

MOCKGEN := $(shell go env GOPATH)/bin/mockgen
mocks: $(MOCKGEN)
	go generate ./pkg/stores/... ./pkg/logger/... ./pkg/validator/... ./pkg/services/... ./pkg/auth/... ./pkg/worker/stack/... ./pkg/worker/release/...
.PHONY: mocks

$(MOCKGEN):
	go install go.uber.org/mock/mockgen@v0.6.0

binary:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/stackdome-server cmd/main.go
.PHONY: binary

installer: ## Build the VPS installer binary
	CGO_ENABLED=0 go build -o bin/stackdome-install ./cmd/installer/
.PHONY: installer

installer-linux-amd64: ## Build installer for linux/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/stackdome-install-linux-amd64 ./cmd/installer/
.PHONY: installer-linux-amd64

installer-linux-arm64: ## Build installer for linux/arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/stackdome-install-linux-arm64 ./cmd/installer/
.PHONY: installer-linux-arm64

all: frontend binary
.PHONY: all

image-build: frontend ## Build API server container image
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/linux_amd64/stackdome-server cmd/main.go
	$(DOCKER) build -t $(IMAGE_TAG) .
	@echo "Built image: $(IMAGE_TAG)"
.PHONY: image-build

image-push: image-build ## Build and push API server container image
	$(DOCKER) push $(IMAGE_TAG)
	@echo "Pushed image: $(IMAGE_TAG)"
.PHONY: image-push

test:
	mage test:unit
.PHONY: test

PG_CONTAINER := psql-stackdome-dev
PG_USER := postgres
PG_PASSWORD := foobar-bizz-buzz
PG_DB := stackdome_dev
PG_PORT := 5432

.PHONY: ensure-postgres
ensure-postgres: SHELL := /usr/bin/env bash
ensure-postgres: ## Start the dev PostgreSQL container if it isn't already running.
	@if pg_isready -h localhost -p $(PG_PORT) -q 2>/dev/null; then \
		echo "PostgreSQL already listening on port $(PG_PORT), skipping container setup."; \
	else \
		if $(DOCKER) ps --format '{{.Names}}' | grep -q '^$(PG_CONTAINER)$$'; then \
			echo "PostgreSQL container '$(PG_CONTAINER)' already running."; \
		elif $(DOCKER) ps -a --format '{{.Names}}' | grep -q '^$(PG_CONTAINER)$$'; then \
			echo "Starting existing PostgreSQL container '$(PG_CONTAINER)'..."; \
			$(DOCKER) start $(PG_CONTAINER); \
		else \
			echo "Creating PostgreSQL container '$(PG_CONTAINER)'..."; \
			$(DOCKER) run --name $(PG_CONTAINER) \
				-e POSTGRES_USER=$(PG_USER) \
				-e POSTGRES_PASSWORD=$(PG_PASSWORD) \
				-e POSTGRES_DB=$(PG_DB) \
				-p $(PG_PORT):5432 \
				-d postgres; \
		fi; \
		echo "Waiting for PostgreSQL to be ready..."; \
		for i in $$(seq 1 30); do \
			$(DOCKER) exec $(PG_CONTAINER) pg_isready -U $(PG_USER) >/dev/null 2>&1 && \
				echo "PostgreSQL ready (port: $(PG_PORT))" && exit 0; \
			sleep 1; \
		done; \
		echo "PostgreSQL failed to start within 30 seconds" && exit 1; \
	fi

.PHONY: test-integration
test-integration: SHELL := /usr/bin/env bash
test-integration: ensure-postgres ## Run integration tests. Optional: FOCUS="My Test Name" to run a specific spec.
	@go test -c -o test/int/integration.test ./test/int
	@cd test/int && set -o pipefail && ./integration.test -test.v -ginkgo.v -test.timeout 30m -test.count 1 \
		$(if $(FOCUS),-ginkgo.focus="$(FOCUS)") \
		2>&1 | tee last-run.log; \
		EXIT_CODE=$$?; rm -f integration.test; exit $$EXIT_CODE
