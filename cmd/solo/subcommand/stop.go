package subcommand

import (
	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/host/app"
	"github.com/spaulg/solo/internal/pkg/host/app/context"
)

func NewStopCommand(soloCtx *context.CliContext) *cobra.Command {
	var profiles []string

	stopCmd := &cobra.Command{
		Use:     "stop",
		GroupID: "lifecycle",
		Short:   "Stops your app",
		Long:    "Stops your app",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return soloCtx.Project.ReloadWithProfiles(profiles)
		},
		RunE: soloCtx.ProtectWithLock(func(_ *cobra.Command, _ []string) error {
			projectControl, err := app.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Stop()
		}),
	}

	stopCmd.Flags().StringSliceVarP(&profiles, "profile", "", nil, "Profiles to use for the command.")

	return stopCmd
}
