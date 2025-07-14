package subcommand

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewShCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "sh [service]",
		GroupID: "tooling",
		Short:   "Drops into a shell on a service",
		Long:    "Drops into a shell on a service",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("service name is required")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			projectControl, err := host.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.ExecuteShell(args[0])
		},
	}
}
