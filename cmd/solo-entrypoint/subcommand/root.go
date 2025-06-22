package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/context"
	"github.com/spf13/cobra"
	"os"
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

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
