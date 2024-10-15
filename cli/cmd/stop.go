package cmd

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectControl := solo.NewProjectControl(config, project)
		return projectControl.Stop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
