package subcommand

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/app"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
)

func NewRebuildCommand(soloCtx *context.CliContext) *cobra.Command {
	var rebuildCmdYes bool

	rebuildCmd := &cobra.Command{
		Use:     "rebuild",
		GroupID: "lifecycle",
		Short:   "Rebuilds your app from scratch, preserving data",
		Long:    "Rebuilds your app from scratch, preserving data",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if err := soloCtx.Project.ReloadWithProfiles([]string{"*"}); err != nil {
				return err
			}

			if !rebuildCmdYes {
				var cmdConfirmationString string
				for {
					fmt.Print("Are you sure you want to rebuild (y/n)? ")
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
		RunE: soloCtx.ProtectWithLock(func(_ *cobra.Command, _ []string) error {
			projectControl, err := app.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Rebuild()
		}),
	}

	rebuildCmd.Flags().BoolVarP(&rebuildCmdYes, "yes", "y", false, "Answer yes non-interactively to confirmation questions")

	return rebuildCmd
}
