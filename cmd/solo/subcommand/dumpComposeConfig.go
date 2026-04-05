package subcommand

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/app/events"
	"github.com/spaulg/solo/internal/pkg/host/infra/container"
)

func NewDumpComposeConfigCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "dump-compose-config",
		GroupID: "config",
		Short:   "Dumps the compose config to stdout",
		Long:    "Dumps the compose config to stdout",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		RunE: func(_ *cobra.Command, _ []string) error {
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
