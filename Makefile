DOCKER ?= docker
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
VERSION := $(shell date +%s)

# Tag for the image
IMAGE_TAG := $(VERSION)

generate:
	rm -rf pkg/api/openapi
	$(DOCKER) run -v ${PWD}:/local:rw openapitools/openapi-generator-cli:v6.0.1 generate -i /local/config/openapi/stackdome_api.yaml -g go -o /local/pkg/api/openapi
	gofmt -w pkg/api/openapi
	rm pkg/api/openapi/go.mod
	rm pkg/api/openapi/go.sum
.PHONY: generate

binary:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o bin/stackdome-server cmd/main.go
.PHONY: binary

image: GOOS=linux
image: binary
	@if [ -z "$(EXTERNAL_IMAGE_REGISTRY)" ]; then \
	  echo "Error: EXTERNAL_IMAGE_REGISTRY is not set"; \
	  exit 1; \
	fi
	@if [ -z "$(IMAGE_REPOSITORY)" ]; then \
	  echo "Error: IMAGE_REPOSITORY is not set"; \
	  exit 1; \
	fi
	$(DOCKER) build -t "$(EXTERNAL_IMAGE_REGISTRY)/$(IMAGE_REPOSITORY):$(IMAGE_TAG)" .
.PHONY: image
