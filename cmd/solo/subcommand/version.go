package subcommand

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewVersionCommand(_ *ProjectConfigContext) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Displays the solo version",
		Long:  "Displays the solo version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("version called")
		},
	}
}
