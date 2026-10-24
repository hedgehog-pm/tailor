package codegen

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/ettle/strcase"
)

func TypeName(slug string) (string, error) {
	name := strcase.ToGoPascal(slug)
	if name != "" && !ast.IsExported(name) {
		name = "Schema" + name
	}

	if !token.IsIdentifier(name) {
		return "", fmt.Errorf("cannot derive Go type name from slug %q", slug)
	}

	return name, nil
}
