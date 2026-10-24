package output

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/renameio/v2"

	"github.com/hedgehog-pm/tailor/internal/codegen"
)

func Write(ctx context.Context, results []codegen.Result) error {
	for _, result := range results {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(result.Path), 0o755); err != nil {
			return fmt.Errorf("create output directory: %w", err)
		}

		if err := renameio.WriteFile(
			result.Path,
			result.Data,
			0o644,
		); err != nil {
			return fmt.Errorf("write output %s: %w", result.Path, err)
		}
	}

	return nil
}
