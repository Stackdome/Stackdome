DOCKER ?= docker
OPENAPI_GENERATOR_IMAGE ?= openapitools/openapi-generator-cli:v6.0.1
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

PROMETHEUS_VERSION ?= 3.13.2
PROMETHEUS_RELEASE_URL := https://github.com/prometheus/prometheus/releases/download/v$(PROMETHEUS_VERSION)
PROMTOOL := bin/promtool
PROMTOOL_OS ?= $(shell go env GOHOSTOS)
PROMTOOL_ARCH ?= $(shell go env GOHOSTARCH)

# Image config
IMAGE_REPO ?= quay.io/stackdome/stackdome
VERSION ?= $(shell git rev-parse --short HEAD)
IMAGE_TAG ?= $(IMAGE_REPO):$(VERSION)

generate:
	rm -rf pkg/api/openapi
	# Run as the invoking user so generated files are not root-owned on Linux
	# (root-owned output breaks the gofmt step and the CI regen check).
	$(DOCKER) run --user $(shell id -u):$(shell id -g) -v ${PWD}:/local:rw $(OPENAPI_GENERATOR_IMAGE) generate -i /local/config/openapi/stackdome_api.yaml -g go -o /local/pkg/api/openapi
	gofmt -w pkg/api/openapi
	rm pkg/api/openapi/go.mod
	rm pkg/api/openapi/go.sum
.PHONY: generate

docs-openapi:
	cp pkg/api/openapi/api/openapi.yaml docs/openapi.yaml
.PHONY: docs-openapi

docs-openapi-source-check:
	@$(DOCKER) run --rm -v ${PWD}:/local:ro --entrypoint /bin/sh $(OPENAPI_GENERATOR_IMAGE) -c '\
		docker-entrypoint.sh generate -i /local/config/openapi/stackdome_api.yaml -g go -o /tmp/stackdome-openapi >/dev/null && \
		cmp -s /tmp/stackdome-openapi/api/openapi.yaml /local/pkg/api/openapi/api/openapi.yaml' || \
		(echo "pkg/api/openapi is stale; run make generate" && exit 1)
.PHONY: docs-openapi-source-check

docs-openapi-check: docs-openapi-source-check
	cmp -s pkg/api/openapi/api/openapi.yaml docs/openapi.yaml || (echo "docs/openapi.yaml is stale; run make docs-openapi" && exit 1)
.PHONY: docs-openapi-check

frontend:
	corepack enable pnpm
	corepack prepare pnpm@10.33.2 --activate
	pnpm --prefix frontend install --frozen-lockfile
	pnpm --prefix frontend exec vite build
	touch pkg/web/dist/.gitkeep
.PHONY: frontend

MOCKGEN := $(shell go env GOPATH)/bin/mockgen
mocks: $(MOCKGEN)
	go generate ./pkg/controllers ./pkg/stores/... ./pkg/logger/... ./pkg/validator/... ./pkg/services/... ./pkg/auth/... ./pkg/worker/stack/... ./pkg/worker/release/... ./pkg/clients/... ./pkg/credentials/... ./pkg/clustermanager/... ./pkg/handlers/... ./pkg/resourceaccess/...
.PHONY: mocks

$(MOCKGEN):
	go install go.uber.org/mock/mockgen@v0.6.0

fmt:
	gofmt -w .
.PHONY: fmt

lint: ## Run golangci-lint (installs pinned version if needed)
	mage lint
.PHONY: lint

promtool: ## Install the pinned promtool release into bin/
	@set -eu; \
	case "$(PROMTOOL_OS)/$(PROMTOOL_ARCH)" in \
		darwin/amd64|darwin/arm64|linux/amd64|linux/arm64) ;; \
		*) echo "unsupported promtool platform: $(PROMTOOL_OS)/$(PROMTOOL_ARCH)" >&2; exit 1 ;; \
	esac; \
	if [ -x "$(PROMTOOL)" ] && "$(PROMTOOL)" --version 2>&1 | grep -q "version $(PROMETHEUS_VERSION)"; then \
		echo "promtool $(PROMETHEUS_VERSION) already installed at $(PROMTOOL)"; \
		exit 0; \
	fi; \
	archive="prometheus-$(PROMETHEUS_VERSION).$(PROMTOOL_OS)-$(PROMTOOL_ARCH).tar.gz"; \
	tmp_dir=$$(mktemp -d); \
	trap 'rm -rf "$$tmp_dir"' EXIT; \
	echo "Installing promtool $(PROMETHEUS_VERSION) to $(PROMTOOL)..."; \
	curl --fail --location --silent --show-error \
		-o "$$tmp_dir/$$archive" "$(PROMETHEUS_RELEASE_URL)/$$archive"; \
	curl --fail --location --silent --show-error \
		-o "$$tmp_dir/sha256sums.txt" "$(PROMETHEUS_RELEASE_URL)/sha256sums.txt"; \
	expected=$$(awk -v archive="$$archive" '$$2 == archive { print $$1 }' "$$tmp_dir/sha256sums.txt"); \
	[ -n "$$expected" ] || { echo "checksum not found for $$archive" >&2; exit 1; }; \
	if command -v sha256sum >/dev/null 2>&1; then \
		actual=$$(sha256sum "$$tmp_dir/$$archive" | awk '{ print $$1 }'); \
	else \
		actual=$$(shasum -a 256 "$$tmp_dir/$$archive" | awk '{ print $$1 }'); \
	fi; \
	[ "$$actual" = "$$expected" ] || { echo "checksum mismatch for $$archive" >&2; exit 1; }; \
	tar -xzf "$$tmp_dir/$$archive" -C "$$tmp_dir"; \
	mkdir -p "$(dir $(PROMTOOL))"; \
	cp "$$tmp_dir/prometheus-$(PROMETHEUS_VERSION).$(PROMTOOL_OS)-$(PROMTOOL_ARCH)/promtool" "$(PROMTOOL)"; \
	chmod 0755 "$(PROMTOOL)"; \
	echo "Installed $(PROMTOOL)"
.PHONY: promtool

observability-check: promtool ## Validate the alpha dashboard and Prometheus rules
	@set -eu; \
	tmp_rules=$$(mktemp); \
	trap 'rm -f "$$tmp_rules"' EXIT; \
	sed -n '/^spec:$$/,$$p' deploy/observability/prometheus-rules.yaml | \
		tail -n +2 | sed 's/^  //' > "$$tmp_rules"; \
	"$(PROMTOOL)" check rules "$$tmp_rules"; \
	jq empty deploy/observability/grafana/alpha-overview.json; \
	echo "Observability artifacts are valid"
.PHONY: observability-check

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
	$(DOCKER) build --platform linux/amd64 -t $(IMAGE_TAG) .
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
	@cd test/int && set -o pipefail && ./integration.test -test.v -ginkgo.v -test.timeout 1h -test.count 1 \
		$(if $(FOCUS),-ginkgo.focus="$(FOCUS)") \
		2>&1 | tee last-run.log; \
		EXIT_CODE=$$?; rm -f integration.test; exit $$EXIT_CODE

.PHONY: test-cloud-integration
test-cloud-integration: SHELL := /usr/bin/env bash
test-cloud-integration: ensure-postgres ## Run the focused Stackdome Cloud/shared integration suite.
	@go test -tags=cloud_e2e -c -o test/cloudint/cloud-integration.test ./test/cloudint
	@cd test/cloudint && set -o pipefail && ./cloud-integration.test -test.v -ginkgo.v -test.timeout 1h -test.count 1 \
		2>&1 | tee last-run.log; \
		EXIT_CODE=$$?; rm -f cloud-integration.test; exit $$EXIT_CODE
