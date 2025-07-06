package subcommand

import (
	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewRestartCommand(soloCtx *context.CliContext) *cobra.Command {
	var profiles []string

	restartCmd := &cobra.Command{
		Use:     "restart",
		GroupID: "lifecycle",
		Short:   "Restarts your app",
		Long:    "Restarts your app",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := loadProjectE(soloCtx, profiles); err != nil {
				return err
			}

			return nil
		},
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

	restartCmd.Flags().StringSliceVarP(&profiles, "profile", "", []string{"*"}, "Profiles to use for the command.")

	return restartCmd
}
