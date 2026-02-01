package subcommand

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewShCommand(soloCtx *context.CliContext) *cobra.Command {
	var shellPath string
	var replica int

	shCmd := &cobra.Command{
		Use:     "sh [service]",
		GroupID: "tooling",
		Short:   "Start a shell in a service",
		Long:    "Start a shell in a service",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("service name is required")
			}

			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			projectControl, err := host.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.ExecuteShell(shellPath, replica, args[0])
		},
	}

	shCmd.Flags().StringVarP(&shellPath, "shell", "s", "", "Override the shell")
	shCmd.Flags().IntVarP(&replica, "replica", "r", 1, "Replica number to target (default: 1)")

	return shCmd
}
