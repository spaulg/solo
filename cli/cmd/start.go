package cmd

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts your app",
	Long:  "Starts your app",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectControl := solo.NewProjectControl(config, project)
		return projectControl.Start()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
