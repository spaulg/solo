package subcommand

import (
	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewStopCommand(soloCtx *context.CliContext) *cobra.Command {
	var profiles []string

	stopCmd := &cobra.Command{
		Use:     "stop",
		GroupID: "lifecycle",
		Short:   "Stops your app",
		Long:    "Stops your app",
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

			return projectControl.Stop()
		}),
	}

	stopCmd.Flags().StringSliceVarP(&profiles, "profile", "", []string{"*"}, "Profiles to use for the command.")

	return stopCmd
}
