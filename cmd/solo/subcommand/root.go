package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/common/logging"
	"github.com/spaulg/solo/internal/pkg/solo/config"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
)

func NewRootCommand(soloCtx *context.SoloContext) *cobra.Command {
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

func Execute() {
	cobra.EnableCommandSorting = false

	soloCtx := loadConfigAndProjectContext()

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

func buildLogHandler(soloCtx *context.SoloContext) (slog.Handler, error) {
	stateDirectory := path.Join(soloCtx.Project.GetStateDirectoryRoot(), "logs")
	if err := os.MkdirAll(stateDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	logFileName := path.Join(stateDirectory, time.Now().Format("cli-2006-01-02.log"))

	logFile, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	switch soloCtx.Config.Logging.Handler {
	case "json":
		return slog.NewJSONHandler(logFile, &slog.HandlerOptions{
			Level: buildLogLevel(soloCtx),
		}), nil

	case "text":
		fallthrough
	default:
		return slog.NewTextHandler(logFile, &slog.HandlerOptions{
			Level: buildLogLevel(soloCtx),
		}), nil
	}
}

func buildLogLevel(soloCtx *context.SoloContext) slog.Level {
	switch soloCtx.Config.Logging.Level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "error":
		return slog.LevelError
	case "warning":
		fallthrough
	default:
		return slog.LevelWarn
	}
}

func loadConfigAndProjectContext() *context.SoloContext {
	loadedConfig, configLoadErr := config.NewConfig()
	loadedProject, projectLoadErr := project.FindProject("./")

	if loadedProject != nil && configLoadErr == nil {
		configLoadErr = loadedConfig.AddConfigPath(loadedProject.GetDirectory())
	}

	return &context.SoloContext{
		Config:         loadedConfig,
		ConfigLoadErr:  configLoadErr,
		Project:        loadedProject,
		ProjectLoadErr: projectLoadErr,
		Logger:         slog.New(logging.NewBlackHoleHandler()),
	}
}
