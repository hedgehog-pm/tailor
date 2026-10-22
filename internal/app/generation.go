package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hedgehog-pm/tailor/internal/codegen"
	"github.com/hedgehog-pm/tailor/internal/config"
	"github.com/hedgehog-pm/tailor/internal/generated/schemas"
	"github.com/hedgehog-pm/tailor/internal/schema"
)

// Run получает схему и собирает код для выбранных полей в памяти.
// Файлы и каталоги при этом не создаёт и не изменяет.
func Run(ctx context.Context, cfg config.Config) (codegen.Result, error) {
	client, err := schemas.NewClientWithResponses(
		cfg.URL,
		schemas.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	)
	if err != nil {
		return codegen.Result{}, err
	}

	response, err := client.GetSchemaWithResponse(
		ctx,
		cfg.Slug,
		cfg.Version,
	)
	if err != nil {
		return codegen.Result{}, fmt.Errorf(
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

		return codegen.Result{}, fmt.Errorf(
			"get schema %s version %d: HTTP %d: %s",
			cfg.Slug,
			cfg.Version,
			response.StatusCode(),
			message,
		)
	}

	projected, err := schema.Project(ctx, *response.JSON200, cfg.Fields)
	if err != nil {
		return codegen.Result{}, fmt.Errorf("project schema: %w", err)
	}

	source, err := codegen.Generate(
		projected,
		cfg.Slug,
		cfg.Package,
	)
	if err != nil {
		return codegen.Result{}, fmt.Errorf("generate Go types: %w", err)
	}

	// Генератор не принимает контекст. Проверяем отмену после его вызова,
	// чтобы не выдавать результат отменённого запуска.
	if err := ctx.Err(); err != nil {
		return codegen.Result{}, err
	}

	return codegen.Result{Path: cfg.Out, Data: source}, nil
}
