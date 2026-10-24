package codegen

import (
	"go/format"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	oapi "github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"go.yaml.in/yaml/v3"

	tasksapi "github.com/hedgehog-pm/tailor/api/clients/tasks"
)

func GenerateHTTPClient(packageName string) ([]byte, error) {
	spec, err := openapi3.NewLoader().LoadFromData(tasksapi.OpenAPI)
	if err != nil {
		return nil, err
	}

	var options oapi.Configuration
	if err := yaml.Unmarshal(tasksapi.Codegen, &options); err != nil {
		return nil, err
	}

	options.PackageName = packageName

	source, err := oapi.Generate(spec, options)
	if err != nil {
		return nil, err
	}

	// oapi-codegen переименовывает тип клиента, но оставляет конструктор NewClient.
	// Это имя нужно нашей обёртке в том же пакете.
	source = strings.ReplaceAll(source, "NewClient(", "newHTTPClient(")

	return format.Source([]byte(source))
}
