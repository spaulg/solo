package subcommand

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/host/app"
	"github.com/spaulg/solo/internal/pkg/host/app/context"
)

func NewDestroySubCommand(soloCtx *context.CliContext) *cobra.Command {
	var destroyCmdYes bool
	var profiles []string

	destroyCmd := &cobra.Command{
		Use:     "destroy",
		GroupID: "lifecycle",
		Short:   "Destroys your app",
		Long:    "Destroys your app",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := soloCtx.Project.ReloadWithProfiles(profiles); err != nil {
				return err
			}

			if !destroyCmdYes {
				var cmdConfirmationString string
				for {
					fmt.Print("Are you sure you want to destroy (y/n)? ")
					_, err := fmt.Scanln(&cmdConfirmationString)

					if err != nil {
						continue
					} else if strings.ToLower(cmdConfirmationString) == "n" {
						return ErrUserAbortedCommand
					} else if strings.ToLower(cmdConfirmationString) == "y" {
						break
					}
				}
			}

			return nil
		},
		RunE: soloCtx.ProtectWithLock(func(_ *cobra.Command, _ []string) error {
			projectControl, err := app.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Destroy()
		}),
	}

	destroyCmd.Flags().BoolVarP(&destroyCmdYes, "yes", "y", false, "Answer yes non-interactively to confirmation questions")
	destroyCmd.Flags().StringSliceVarP(&profiles, "profile", "", []string{"*"}, "Profiles to use for the command.")

	return destroyCmd
}
