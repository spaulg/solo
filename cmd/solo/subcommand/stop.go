package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spf13/cobra"
)

func NewStopCommand(ctx *ProjectConfigContext) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stops your app",
		Long:  "Stops your app",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(ctx.Config, ctx.Project)
			if err != nil {
				return err
			}

			return projectControl.Stop()
		},
	}
}
