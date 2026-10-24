package codegen

import (
	"go/format"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	oapi "github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"

	"github.com/hedgehog-pm/tailor/api"
)

func GenerateHTTPClient(packageName string) ([]byte, error) {
	spec, err := openapi3.NewLoader().LoadFromData(api.Tasks)
	if err != nil {
		return nil, err
	}

	source, err := oapi.Generate(spec, oapi.Configuration{
		PackageName: packageName,
		Generate:    oapi.GenerateOptions{Models: true, Client: true},
		OutputOptions: oapi.OutputOptions{
			ClientTypeName:      "HTTPClient",
			IncludeOperationIDs: []string{"getTask"},
		},
	})
	if err != nil {
		return nil, err
	}

	// oapi-codegen переименовывает тип клиента, но оставляет конструктор NewClient.
	// Это имя нужно нашей обёртке в том же пакете.
	source = strings.ReplaceAll(source, "NewClient(", "newHTTPClient(")

	return format.Source([]byte(source))
}
