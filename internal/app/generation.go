package app

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	jsonschema "github.com/hedgehog-pm/schema"

	"github.com/hedgehog-pm/tailor/internal/codegen"
	"github.com/hedgehog-pm/tailor/internal/config"
	schemas "github.com/hedgehog-pm/tailor/internal/generated/clients/taskschemas"
	"github.com/hedgehog-pm/tailor/internal/schema"
)

// Run получает схему и собирает код для выбранных полей в памяти.
// Файлы и каталоги при этом не создаёт и не изменяет.
func Run(ctx context.Context, cfg config.Config) ([]codegen.Result, error) {
	client, err := schemas.NewClientWithResponses(
		cfg.URL,
		schemas.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	)
	if err != nil {
		return nil, err
	}

	response, err := client.GetSchemaWithResponse(
		ctx,
		cfg.Slug,
		cfg.Version,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get schema %s version %d: %w",
			cfg.Slug,
			cfg.Version,
			err,
		)
	}

	if response.JSON200 == nil {
		message := response.Status()
		for _, problem := range []*schemas.Error{response.JSON400, response.JSON404, response.JSON500} {
			if problem != nil {
				message = problem.Message
				break
			}
		}

		return nil, fmt.Errorf(
			"get schema %s version %d: HTTP %d: %s",
			cfg.Slug,
			cfg.Version,
			response.StatusCode(),
			message,
		)
	}

	compiler, err := jsonschema.New(1)
	if err != nil {
		return nil, err
	}

	document, err := compiler.Compile(*response.JSON200)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}

	projected, err := schema.Project(ctx, document, cfg.Fields)
	if err != nil {
		return nil, fmt.Errorf("project schema: %w", err)
	}

	typeName, err := codegen.TypeName(cfg.Slug)
	if err != nil {
		return nil, err
	}

	source, err := codegen.Generate(
		projected,
		typeName,
		cfg.Package,
	)
	if err != nil {
		return nil, fmt.Errorf("generate Go types: %w", err)
	}

	clientSource, err := codegen.GenerateClient(
		cfg.Package,
		typeName,
		fmt.Sprintf("%s@%d", cfg.Slug, cfg.Version),
	)
	if err != nil {
		return nil, fmt.Errorf("generate Go client: %w", err)
	}

	httpSource, err := codegen.GenerateHTTPClient(cfg.Package)
	if err != nil {
		return nil, fmt.Errorf("generate HTTP client: %w", err)
	}

	// Генератор не принимает контекст. Проверяем отмену после его вызова,
	// чтобы не выдавать результат отменённого запуска.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return []codegen.Result{
		{Path: cfg.Out, Data: source},
		{
			Path: filepath.Join(filepath.Dir(cfg.Out), "client.go"),
			Data: clientSource,
		},
		{
			Path: filepath.Join(filepath.Dir(cfg.Out), "http_client.go"),
			Data: httpSource,
		},
	}, nil
}
