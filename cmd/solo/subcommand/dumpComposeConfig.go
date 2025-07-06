package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/container"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/events"
)

func NewDumpComposeConfigCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:         "dump-compose-config",
		GroupID:     "config",
		Short:       "Dumps the compose config to stdout",
		Long:        "Dumps the compose config to stdout",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := loadProjectE(soloCtx, []string{}); err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			eventManager := events.GetEventManagerInstance()
			orchestrator, err := container.NewOrchestratorFactory(soloCtx, eventManager).Build()

			if err != nil {
				return err
			}

			composeYml, err := orchestrator.ExportComposeConfiguration(soloCtx.Config, soloCtx.Project)

			if err != nil {
				return err
			}

			fmt.Println(string(composeYml))
			return nil
		},
	}
}
