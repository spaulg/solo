package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spf13/cobra"
)

func NewStopCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:         "stop",
		GroupID:     "lifecycle",
		Short:       "Stops your app",
		Long:        "Stops your app",
		Annotations: map[string]string{LoadProjectFileAnnotation: "true"},
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			projectControl, err := host.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Stop()
		}),
	}
}
