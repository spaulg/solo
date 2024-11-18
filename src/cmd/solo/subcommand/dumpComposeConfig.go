package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/orchestrator"
	"github.com/spf13/cobra"
)

var composeConfigCmd = &cobra.Command{
	Use:   "dump-compose-config",
	Short: "Dumps the compose config to stdout",
	Long:  "Dumps the compose config to stdout",
	RunE: func(cmd *cobra.Command, args []string) error {
		composeYml, err := orchestrator.OrchestratorFactory(config).ExportComposeConfiguration(config, project)
		if err != nil {
			return err
		}

		fmt.Println(string(composeYml))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(composeConfigCmd)
}
