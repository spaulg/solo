package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func NewRebuildCommand(soloCtx *context.CliContext) *cobra.Command {
	var rebuildCmdYes bool

	rebuildCmd := &cobra.Command{
		Use:     "rebuild",
		GroupID: "lifecycle",
		Short:   "Rebuilds your app from scratch, preserving data",
		Long:    "Rebuilds your app from scratch, preserving data",
		PreRunE: func(cmd *cobra.Command, args []string) error {
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
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			if err := projectControl.Destroy(); err != nil {
				return err
			}

			if err := projectControl.Clean(false); err != nil {
				return err
			}

			return projectControl.Start()
		}),
	}

	rebuildCmd.Flags().BoolVarP(&rebuildCmdYes, "yes", "y", false, "Answer yes non-interactively to confirmation questions")

	return rebuildCmd
}
