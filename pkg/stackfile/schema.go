package stackfile

import _ "embed"

// SchemaJSON is the JSON Schema (draft-07) for stackfile documents.
//
//go:embed schema.json
var SchemaJSON []byte
