package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spf13/cobra"
)

func NewStartCommand(ctx *ProjectConfigContext) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Starts your app",
		Long:  "Starts your app",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(ctx.Config, ctx.Project)
			if err != nil {
				return err
			}

			return projectControl.Start()
		},
	}
}
