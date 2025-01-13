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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := soloCtx.TryLock(); err != nil {
				return err
			}

			if !destroyCmdForce {
				var destroyCmdForceString string
				for {
					fmt.Print("Are you sure you want to destroy (y/n)? ")
					_, err := fmt.Scanln(&destroyCmdForceString)

					if err != nil {
						continue
					} else if strings.ToLower(destroyCmdForceString) == "n" {
						_ = soloCtx.Unlock()
						os.Exit(0)
					} else if strings.ToLower(destroyCmdForceString) == "y" {
						break
					}
				}
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			if err := projectControl.Destroy(); err != nil {
				return err
			}

			return projectControl.Clean(false)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			return soloCtx.Unlock()
		},
	}

	destroyCmd.Flags().BoolVarP(&destroyCmdForce, "force", "f", false, "Force execution")

	return destroyCmd
}
