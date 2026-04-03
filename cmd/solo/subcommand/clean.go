package subcommand

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/app"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
)

func NewCleanSubCommand(soloCtx *context.CliContext) *cobra.Command {
	var cleanCmdYes bool

	cleanCmd := &cobra.Command{
		Use:     "clean",
		GroupID: "lifecycle",
		Short:   "Clean the app",
		Long:    "Clean the app",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := soloCtx.Project.ReloadWithProfiles([]string{"*"}); err != nil {
				return err
			}

			if !cleanCmdYes {
				var cmdConfirmationString string
				for {
					fmt.Print("Are you sure you want to clean all state data (including a destruction of the app if running) (y/n)? ")
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

			return projectControl.Clean(true)
		}),
	}

	cleanCmd.Flags().BoolVarP(&cleanCmdYes, "yes", "y", false, "Answer yes non-interactively to confirmation questions")

	return cleanCmd
}
