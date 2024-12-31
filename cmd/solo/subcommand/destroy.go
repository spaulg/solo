package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func NewDestroySubCommand(soloCtx *context.SoloContext) *cobra.Command {
	var destroyCmdForce bool

	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroys your app",
		Long:  "Destroys your app",
		PreRun: func(cmd *cobra.Command, args []string) {
			if !destroyCmdForce {
				var destroyCmdForceString string
				for {
					fmt.Print("Are you sure you want to destroy (y/n)? ")
					_, err := fmt.Scanln(&destroyCmdForceString)

					if err != nil {
						continue
					} else if strings.ToLower(destroyCmdForceString) == "n" {
						os.Exit(0)
					} else if strings.ToLower(destroyCmdForceString) == "y" {
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

			return projectControl.Destroy(true)
		},
	}

	destroyCmd.Flags().BoolVarP(&destroyCmdForce, "force", "f", false, "Force execution")

	return destroyCmd
}
