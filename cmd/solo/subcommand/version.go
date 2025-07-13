package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewVersionCommand(_ *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Displays the solo version",
		Long:  "Displays the solo version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("version called")
		},
	}
}
