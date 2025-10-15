package subcommand

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
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
						os.Exit(0)
					} else if strings.ToLower(cmdConfirmationString) == "y" {
						break
					}
				}
			}

			return nil
		},
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			projectControl, err := host.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Clean(true)
		}),
	}

	cleanCmd.Flags().BoolVarP(&cleanCmdYes, "yes", "y", false, "Answer yes non-interactively to confirmation questions")

	return cleanCmd
}
