package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewLogsCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "logs",
		GroupID: "tooling",
		Short:   "Displays logs for your app",
		Long:    "Displays logs for your app",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := loadProjectE(soloCtx, []string{}); err != nil {
				return err
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("logs called")
		},
	}
}
