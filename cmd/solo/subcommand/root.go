package subcommand

import (
	"fmt"
	"os"
	"strings"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"

	"github.com/spf13/cobra"
)

const RequireConfigLoadSuccessAnnotation = "RequireConfigLoadSuccess"
const RequireProjectLoadSuccessAnnotation = "RequireProjectLoadSuccess"

func Execute() {
	cobra.EnableCommandSorting = false

	soloCtx, err := context.LoadCliContext()
	if err != nil {
		os.Exit(1)
	}

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
	rootCmd.AddCommand(NewShCommand(soloCtx))
	rootCmd.AddCommand(NewLogsCommand(soloCtx))
	rootCmd.AddCommand(NewToolCommands(soloCtx)...)

	// Config
	rootCmd.AddCommand(NewDumpConfigCommand(soloCtx))
	rootCmd.AddCommand(NewDumpComposeConfigCommand(soloCtx))

	// Other
	rootCmd.AddCommand(NewVersionCommand(soloCtx))

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func NewRootCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:          "solo",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Annotations == nil || cmd.Annotations[RequireConfigLoadSuccessAnnotation] != "true" {
				return nil
			}

			if soloCtx.ConfigLoadErr != nil {
				fmt.Println(soloCtx.ConfigLoadErr)
				os.Exit(1)
			}

			if cmd.Annotations == nil || cmd.Annotations[RequireProjectLoadSuccessAnnotation] != "true" {
				return nil
			}

			if soloCtx.ProjectLoadErr != nil {
				fmt.Println(soloCtx.ProjectLoadErr)
				os.Exit(1)
			}

			soloCtx.CommandPath = strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], " ")
			soloCtx.CommandArgs = os.Args[1:]

			return nil
		},
	}
}
