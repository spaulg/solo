package cmd

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo"
	"github.com/spf13/cobra"
)

// rebuildCmd represents the rebuild command
var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuilds your app from scratch, preserving data",
	Long:  "Rebuilds your app from scratch, preserving data",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectControl := solo.NewProjectControl(config, project)

		if err := projectControl.Destroy(); err != nil {
			return err
		}

		return projectControl.Start()
	},
}

func init() {
	rootCmd.AddCommand(rebuildCmd)
}
