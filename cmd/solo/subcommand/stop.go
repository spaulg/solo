package subcommand

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spaulg/solo/internal/pkg/solo/bubbletea/models"
	"github.com/spaulg/solo/internal/pkg/solo/bubbletea/subscribers"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spf13/cobra"
)

func NewStopCommand(soloCtx *context.CliContext) *cobra.Command {
	return &cobra.Command{
		Use:     "stop",
		GroupID: "lifecycle",
		Short:   "Stops your app",
		Long:    "Stops your app",
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			model, err := models.NewStopModel(soloCtx)
			if err != nil {
				return err
			}

			p := tea.NewProgram(*model)

			eventManager := events.GetEventManagerInstance()
			eventManager.Subscribe(subscribers.NewEventBusToBubbleTeaBridge(soloCtx, p))

			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		}),
	}
}
