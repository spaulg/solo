package cmd

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops your app",
	Long:  "Stops your app",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectControl := solo.ProjectControlFactory(config, project)
		return projectControl.Stop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
