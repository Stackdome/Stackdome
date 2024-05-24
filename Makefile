
DOCKER?=docker

generate:
	rm -rf pkg/api/openapi
	$(DOCKER) run --rm -v ${PWD}:/local:rw openapitools/openapi-generator-cli:v6.0.1 generate -i /local/config/openapi/api-server.yaml -g go -o /local/pkg/api/openapi
	gofmt -w pkg/api/openapi
	rm pkg/api/openapi/go.mod
	rm pkg/api/openapi/go.sum
.PHONY: generate


binary:
	go build -o bin/sordev-server cmd/main.go
.PHONY: binary