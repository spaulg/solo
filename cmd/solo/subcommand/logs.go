package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
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
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println("logs called")
		},
	}
}
