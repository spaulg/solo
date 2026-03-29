package subcommand

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/app/context"
)

func NewRootCommand(_ *context.EntrypointContext) *cobra.Command {
	return &cobra.Command{
		Use:          "solo",
		SilenceUsage: true,
	}
}

func Execute() {
	cobra.EnableCommandSorting = false

	entrypointCtx, err := context.LoadEntrypointContext()
	if err != nil {
		os.Exit(1)
	}

	rootCmd := NewRootCommand(entrypointCtx)
	rootCmd.AddCommand(NewEntrypointCommand(entrypointCtx))
	rootCmd.AddCommand(NewTriggerEventCommand(entrypointCtx))
	rootCmd.AddCommand(NewCatShellsCommand(entrypointCtx))

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
