package cmd

import (
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var destroyCmdForce bool

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroys your app",
	Long:  "Destroys your app",
	PreRun: func(cmd *cobra.Command, args []string) {
		if destroyCmdForce == false {
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
		projectControl, err := solo.ProjectControlFactory(config, project)
		if err != nil {
			return err
		}

		return projectControl.Destroy()
	},
}

func init() {
	destroyCmd.Flags().BoolVarP(&destroyCmdForce, "force", "f", false, "Force execution")
	rootCmd.AddCommand(destroyCmd)
}
