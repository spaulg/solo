package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/container"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
)

func NewDumpComposeConfigCommand(soloCtx *context.SoloContext) *cobra.Command {
	return &cobra.Command{
		Use:     "dump-compose-config",
		GroupID: "config",
		Short:   "Dumps the compose config to stdout",
		Long:    "Dumps the compose config to stdout",
		RunE: func(cmd *cobra.Command, args []string) error {
			composeYml, err := container.NewOrchestratorFactory().Build(soloCtx).
				ExportComposeConfiguration(soloCtx.Config, soloCtx.Project)

			if err != nil {
				return err
			}

			fmt.Println(string(composeYml))
			return nil
		},
	}
}
