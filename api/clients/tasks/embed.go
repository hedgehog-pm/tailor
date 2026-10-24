package tasks

import _ "embed"

//go:embed openapi.yaml
var OpenAPI []byte

//go:embed oapi-codegen.yaml
var Codegen []byte
