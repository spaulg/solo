package subcommand

import (
	"github.com/spf13/cobra"

	tea "charm.land/bubbletea/v2"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/views/workflow_log"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
)

func NewWorkflowLogSubCommand(soloCtx *context.CliContext) *cobra.Command {
	buildLogCmd := &cobra.Command{
		Use:     "wflogs",
		GroupID: "tooling",
		Short:   "Workflow logs",
		Long:    "Workflow logs",
		Aliases: []string{"wflog"},
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			model, err := workflow_log.NewView(soloCtx)
			if err != nil {
				return err
			}

			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		}),
	}

	return buildLogCmd
}
