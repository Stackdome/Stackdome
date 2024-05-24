package openapi

import _ "embed"

//go:embed api-server.yaml
var OpenAPISpec []byte
