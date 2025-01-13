package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
)

func NewRestartCommand(soloCtx *context.SoloContext) *cobra.Command {
	return &cobra.Command{
		Use:     "restart",
		GroupID: "lifecycle",
		Short:   "Restarts your app",
		Long:    "Restarts your app",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return soloCtx.TryLock()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			if err := projectControl.Stop(); err != nil {
				return err
			}

			return projectControl.Start()
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return soloCtx.Unlock()
		},
	}
}
