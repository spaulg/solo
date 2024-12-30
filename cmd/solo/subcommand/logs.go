package subcommand

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewLogsCommand(_ *ProjectConfigContext) *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Displays logs for your app",
		Long:  "Displays logs for your app",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("logs called")
		},
	}
}
