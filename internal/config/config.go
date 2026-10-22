package config

import (
	"errors"
	"fmt"
	"go/build"
	"go/token"
	"os"
	"path/filepath"
)

type Config struct {
	URL     string
	Slug    string
	Version int
	Fields  []string
	Out     string
	Package string
}

// Prepare заполняет пропущенные поля и валидирует конфиг.
func Prepare(cfg Config) (Config, error) {
	if cfg.URL == "" {
		return Config{}, errors.New("SERVICE_TASK_SCHEMAS_URL is not set")
	}

	if cfg.Package == "" {
		name, err := packageName(cfg.Out)
		if err != nil {
			return Config{}, err
		}

		cfg.Package = name
	}

	// IsIdentifier допускает "_", но объявить пакет с таким именем нельзя.
	if !token.IsIdentifier(cfg.Package) || cfg.Package == "_" {
		return Config{}, fmt.Errorf(
			"invalid Go package name %q; specify --package",
			cfg.Package,
		)
	}

	return cfg, nil
}

func packageName(out string) (string, error) {
	dir, err := filepath.Abs(filepath.Dir(out))
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(dir); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return filepath.Base(dir), nil
		}

		return "", err
	}

	pkg, err := build.Default.ImportDir(dir, 0)
	if _, ok := errors.AsType[*build.NoGoError](err); ok {
		return filepath.Base(dir), nil
	}

	if err != nil {
		return "", fmt.Errorf(
			"read output package (or specify --package): %w",
			err,
		)
	}

	return pkg.Name, nil
}
