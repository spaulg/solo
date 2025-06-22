package subcommand

import (
	"fmt"
	"os"
	"strings"

	"github.com/spaulg/solo/internal/pkg/impl/host"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spf13/cobra"
)

func NewDestroySubCommand(soloCtx *context.CliContext) *cobra.Command {
	var destroyCmdYes bool

	destroyCmd := &cobra.Command{
		Use:     "destroy",
		GroupID: "lifecycle",
		Short:   "Destroys your app",
		Long:    "Destroys your app",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !destroyCmdYes {
				var cmdConfirmationString string
				for {
					fmt.Print("Are you sure you want to destroy (y/n)? ")
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

			if err := projectControl.Destroy(); err != nil {
				return err
			}

			return projectControl.Clean(false)
		}),
	}

	destroyCmd.Flags().BoolVarP(&destroyCmdYes, "yes", "y", false, "Answer yes non-interactively to confirmation questions")

	return destroyCmd
}
