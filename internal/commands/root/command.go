package root

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/hedgehog-pm/tailor/internal/app"
	"github.com/hedgehog-pm/tailor/internal/config"
	"github.com/hedgehog-pm/tailor/internal/output"
)

func Execute() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	command, err := NewCommand()
	if err == nil {
		err = command.ExecuteContext(ctx)
	}

	stop()

	if err == nil {
		return
	}

	slog.ErrorContext(ctx, "generation failed", slog.Any("error", err))
	os.Exit(1)
}

func NewCommand() (*cobra.Command, error) {
	var options config.Config

	command := &cobra.Command{
		Use:           "tailor",
		Short:         "Генерация типов по выбранным полям схемы",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(command *cobra.Command, _ []string) error {
			options.URL = os.Getenv("SERVICE_TASK_SCHEMAS_URL")

			cfg, err := config.Prepare(options)
			if err != nil {
				return err
			}

			ctx := command.Context()

			result, err := app.Run(ctx, cfg)
			if err != nil {
				return err
			}

			return output.Write(ctx, result)
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.Slug, "slug", "", "идентификатор схемы")
	flags.IntVar(&options.Version, "version", 0, "версия схемы")
	flags.StringSliceVar(
		&options.Fields,
		"fields",
		nil,
		"имена полей через запятую",
	)
	flags.StringVar(&options.Out, "out", "", "путь к результирующему файлу")
	flags.StringVar(
		&options.Package,
		"package",
		"",
		"имя пакета (по умолчанию — пакет целевого каталога)",
	)

	for _, name := range []string{"slug", "version", "fields", "out"} {
		if err := command.MarkFlagRequired(name); err != nil {
			return nil, err
		}
	}

	return command, nil
}
