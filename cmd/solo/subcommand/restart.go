package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spf13/cobra"
)

func NewRestartCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "restart",
		GroupID: "lifecycle",
		Short:   "Restarts your app",
		Long:    "Restarts your app",
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			projectControl, err := host.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			if err := projectControl.Stop(); err != nil {
				return err
			}

			return projectControl.Start()
		}),
	}
}
