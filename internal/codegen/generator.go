package codegen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atombender/go-jsonschema/pkg/generator"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

type Result struct {
	Path string
	Data []byte
}

// Generate собирает отформатированный код с объявлениями типов из json-схемы,
// не записывая файлы на диск. Любое предупреждение генератора считается ошибкой,
// чтобы не пропустить потерю типизации.
func Generate(
	document *schemas.Schema,
	slug, packageName string,
) ([]byte, error) {
	var warnings []string

	g, err := generator.New(generator.Config{
		DefaultPackageName: packageName,
		DefaultOutputName:  "types.go",
		ResolveExtensions:  []string{".json"},
		Tags:               []string{"json"},
		OnlyModels:         true,
		Warner: func(message string) {
			warnings = append(warnings, message)
		},
	})
	if err != nil {
		return nil, err
	}

	// AddFile требует имя файла как идентификатор схемы. Из него генератор
	// берёт имя для типа, который соответствует всей схеме целиком.
	// Файл при этом не читается, схема уже передана напрямую.
	if err := g.AddFile(slug+".json", document); err != nil {
		return nil, fmt.Errorf("go-jsonschema: %w", err)
	}

	sources, err := g.Sources()
	if err != nil {
		return nil, err
	}

	// Генератор может заменить неподдерживаемый тип на пустой интерфейс.
	// Возвращаем ошибку на любое предупреждение, чтобы не потерять типизацию.
	if len(warnings) > 0 {
		return nil, fmt.Errorf(
			"go-jsonschema: %s",
			strings.Join(warnings, "; "),
		)
	}

	// Проверяем, что по ключу, который мы передали для генерации, действительно
	// есть сгенерированный код.
	source, ok := sources["types.go"]
	if !ok {
		return nil, errors.New("go-jsonschema produced no source")
	}

	return source, nil
}
