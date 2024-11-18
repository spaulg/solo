package cmd

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var rebuildCmdForce bool

var rebuildCmd = &cobra.Command{
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
		projectControl, err := solo.ProjectControlFactory(config, project)
		if err != nil {
			return err
		}

		if err := projectControl.Destroy(); err != nil {
			return err
		}

		return projectControl.Start()
	},
}

func init() {
	rebuildCmd.Flags().BoolVarP(&rebuildCmdForce, "force", "f", false, "Force execution")
	rootCmd.AddCommand(rebuildCmd)
}
