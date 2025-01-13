package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
)

func NewLogsCommand(_ *context.SoloContext) *cobra.Command {
	return &cobra.Command{
		Use:     "logs",
		GroupID: "tooling",
		Short:   "Displays logs for your app",
		Long:    "Displays logs for your app",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("logs called")
		},
	}
}
