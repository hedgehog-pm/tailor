package schema

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/atombender/go-jsonschema/pkg/schemas"
	jsonschema "github.com/hedgehog-pm/schema"
)

// Project возвращает json-схему, в которой от исходной остались только
// перечисленные в fields свойства корневого объекта. Значения этих свойств
// не меняются, включая вложенные объекты, поэтому по результату можно
// сгенерировать их полные Go-типы.
func Project(
	ctx context.Context,
	source *jsonschema.Schema,
	fields []string,
) (*schemas.Schema, error) {
	// Раскрываем ссылки до удаления частей allOf, чтобы сохранить описания,
	// на которые ссылаются выбранные поля.
	inlined, err := source.Inline(ctx)
	if err != nil {
		return nil, err
	}

	document, err := schemas.FromJSONReader(bytes.NewReader(inlined))
	if err != nil {
		return nil, err
	}

	if document.ObjectAsType == nil {
		return nil, errors.New("schema has no root")
	}

	selected := make(map[string]bool, len(fields))
	for _, field := range fields {
		selected[field] = true
	}

	found := make(map[string]bool, len(fields))
	project((*schemas.Type)(document.ObjectAsType), selected, found)

	for _, field := range fields {
		if !found[field] {
			return nil, fmt.Errorf("field %q not found", field)
		}
	}

	// Генератор определяет тип схемы по полю type. Без явного object он
	// пропускает схему целиком, даже если у неё есть properties.
	document.Type = schemas.TypeList{schemas.TypeNameObject}

	return document, nil
}

func project(node *schemas.Type, selected, found map[string]bool) bool {
	maps.DeleteFunc(node.Properties, func(name string, _ *schemas.Type) bool {
		if selected[name] {
			found[name] = true
		}

		return !selected[name]
	})
	// Удаление из properties не снимает обязательность поля в required.
	node.Required = slices.DeleteFunc(node.Required, func(name string) bool {
		return !selected[name]
	})
	node.AllOf = slices.DeleteFunc(node.AllOf, func(part *schemas.Type) bool {
		return !project(part, selected, found)
	})

	return len(node.Properties) > 0 || len(node.AllOf) > 0
}
