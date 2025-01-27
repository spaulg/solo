package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
)

func NewStopCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "stop",
		GroupID: "lifecycle",
		Short:   "Stops your app",
		Long:    "Stops your app",
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Stop()
		}),
	}
}
