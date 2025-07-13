package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewLogsCommand(_ *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "logs",
		GroupID: "tooling",
		Short:   "Displays logs for your app",
		Long:    "Displays logs for your app",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("logs called")
		},
	}
}
