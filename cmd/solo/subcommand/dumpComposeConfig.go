package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/orchestrator"
	"github.com/spf13/cobra"
)

func NewDumpComposeConfigCommand(ctx *ProjectConfigContext) *cobra.Command {
	return &cobra.Command{
		Use:   "dump-compose-config",
		Short: "Dumps the compose config to stdout",
		Long:  "Dumps the compose config to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			composeYml, err := orchestrator.OrchestratorFactory(ctx.Config).
				ExportComposeConfiguration(ctx.Config, ctx.Project)

			if err != nil {
				return err
			}

			fmt.Println(string(composeYml))
			return nil
		},
	}
}
