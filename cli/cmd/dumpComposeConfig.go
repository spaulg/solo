package cmd

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo"
	"github.com/spf13/cobra"
)

// composeConfigCmd represents the composeConfig command
var composeConfigCmd = &cobra.Command{
	Use:   "dump-compose-config",
	Short: "Dumps the compose config to stdout",
	Long:  "Dumps the compose config to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectControl := solo.NewProjectControl(config, project)
		return projectControl.DumpComposeConfig()
	},
}

func init() {
	rootCmd.AddCommand(composeConfigCmd)
}
