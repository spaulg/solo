package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
)

func NewCleanSubCommand(soloCtx *context.SoloContext) *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean the app",
		Long:  "Clean the app",
		RunE: func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Clean(true)
		},
	}
}
