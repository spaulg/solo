package subcommand

import (
	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
)

func NewStartCommand(soloCtx *context.CliContext) *cobra.Command {
	var profiles []string

	startCmd := &cobra.Command{
		Use:         "start",
		GroupID:     "lifecycle",
		Short:       "Starts your app",
		Long:        "Starts your app",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation: "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return soloCtx.Project.ReloadWithProfiles(profiles)
		},
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			projectControl, err := host.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Start()
		}),
	}

	startCmd.Flags().StringSliceVarP(&profiles, "profile", "", nil, "Profiles to use for the command.")

	return startCmd
}
