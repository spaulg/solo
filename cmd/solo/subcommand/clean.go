package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
)

func NewCleanSubCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "clean",
		GroupID: "lifecycle",
		Short:   "Clean the app",
		Long:    "Clean the app",
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			if err := projectControl.Destroy(); err != nil {
				return err
			}

			return projectControl.Clean(true)
		}),
	}
}
