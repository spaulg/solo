package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops your app",
	Long:  "Stops your app",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectControl, err := solo.ProjectControlFactory(config, project)
		if err != nil {
			return err
		}

		return projectControl.Stop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
