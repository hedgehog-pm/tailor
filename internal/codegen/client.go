package codegen

import (
	"bytes"
	"embed"
	"go/format"
	"text/template"
)

//go:embed templates/*.go.tmpl
var templatesFS embed.FS

func GenerateClient(
	packageName, typeName, schemaHeader string,
) ([]byte, error) {
	tmpl, err := template.New("client.go.tmpl").
		ParseFS(templatesFS, "templates/client.go.tmpl")
	if err != nil {
		return nil, err
	}

	data := struct {
		Package      string
		TypeName     string
		SchemaHeader string
	}{Package: packageName, TypeName: typeName, SchemaHeader: schemaHeader}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, data); err != nil {
		return nil, err
	}

	return format.Source(b.Bytes())
}
