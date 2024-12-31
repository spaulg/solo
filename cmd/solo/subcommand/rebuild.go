package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func NewRebuildCommand(soloCtx *context.SoloContext) *cobra.Command {
	var rebuildCmdForce bool

	rebuildCmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuilds your app from scratch, preserving data",
		Long:  "Rebuilds your app from scratch, preserving data",
		PreRun: func(cmd *cobra.Command, args []string) {
			if !rebuildCmdForce {
				var rebuildCmdForceString string
				for {
					fmt.Print("Are you sure you want to rebuild (y/n)? ")
					_, err := fmt.Scanln(&rebuildCmdForceString)

					if err != nil {
						continue
					} else if strings.ToLower(rebuildCmdForceString) == "n" {
						os.Exit(0)
					} else if strings.ToLower(rebuildCmdForceString) == "y" {
						break
					}
				}
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			if err := projectControl.Destroy(false); err != nil {
				return err
			}

			return projectControl.Start()
		},
	}

	rebuildCmd.Flags().BoolVarP(&rebuildCmdForce, "force", "f", false, "Force execution")

	return rebuildCmd
}
