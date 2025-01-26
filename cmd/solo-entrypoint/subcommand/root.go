package subcommand

import (
	"github.com/spf13/cobra"
	"os"
)

func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:          "solo",
		SilenceUsage: true,
	}
}

func Execute() {
	cobra.EnableCommandSorting = false

	rootCmd := NewRootCommand()
	rootCmd.AddCommand(NewEntrypointCommand())
	rootCmd.AddCommand(NewTriggerEventCommand())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
