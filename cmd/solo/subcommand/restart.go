package subcommand

import (
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spf13/cobra"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restarts your app",
	Long:  "Restarts your app",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectControl, err := solo.ProjectControlFactory(config, project)
		if err != nil {
			return err
		}

		if err := projectControl.Stop(); err != nil {
			return err
		}

		return projectControl.Start()
	},
}

func init() {
	rootCmd.AddCommand(restartCmd)
}
