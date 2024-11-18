package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts your app",
	Long:  "Starts your app",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectControl, err := solo.ProjectControlFactory(config, project)
		if err != nil {
			return err
		}

		return projectControl.Start()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
