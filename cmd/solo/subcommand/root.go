package subcommand

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/spaulg/solo/internal/pkg/impl/common/logging"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"

	"github.com/spf13/cobra"
)

func Execute() {
	cobra.EnableCommandSorting = false

	soloCtx := context.LoadCliContext()

	rootCmd := NewRootCommand(soloCtx)
	rootCmd.AddGroup(&cobra.Group{ID: "lifecycle", Title: "Life Cycle Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "tooling", Title: "Tooling Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "config", Title: "Config Commands:"})

	// Lifecycle
	rootCmd.AddCommand(NewStartCommand(soloCtx))
	rootCmd.AddCommand(NewStopCommand(soloCtx))
	rootCmd.AddCommand(NewRestartCommand(soloCtx))
	rootCmd.AddCommand(NewRebuildCommand(soloCtx))
	rootCmd.AddCommand(NewDestroySubCommand(soloCtx))
	rootCmd.AddCommand(NewCleanSubCommand(soloCtx))

	// Tooling
	rootCmd.AddCommand(NewSSHCommand(soloCtx))
	rootCmd.AddCommand(NewLogsCommand(soloCtx))

	// Config
	rootCmd.AddCommand(NewDumpConfigCommand(soloCtx))
	rootCmd.AddCommand(NewDumpComposeConfigCommand(soloCtx))

	// Other
	rootCmd.AddCommand(NewVersionCommand(soloCtx))

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func NewRootCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:          "solo",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if soloCtx.ProjectLoadErr != nil {
				fmt.Println(soloCtx.ProjectLoadErr)
				os.Exit(1)
			}

			if soloCtx.ConfigLoadErr != nil {
				fmt.Println(soloCtx.ConfigLoadErr)
				os.Exit(1)
			}

			// If logging is enabled override the default logger
			if soloCtx.Config.Logging.Enabled {
				handler, err := buildLogHandler(soloCtx)
				if err != nil {
					return err
				}

				soloCtx.Logger = slog.New(handler)
			}

			return nil
		},
	}
}

func buildLogHandler(soloCtx *context.CliContext) (slog.Handler, error) {
	stateDirectory := path.Join(soloCtx.Project.GetStateDirectoryRoot(), "cli", "logs")
	if err := os.MkdirAll(stateDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	logFileName := path.Join(stateDirectory, time.Now().Format("2006-01-02.log"))

	config := soloCtx.Config
	builder := logging.NewLogHandlerBuilder()
	return builder.
		WithLogFilePath(logFileName).
		WithLogLevel(config.Logging.Level).
		WithLogHandlerName(config.Logging.Handler).
		Build()
}
