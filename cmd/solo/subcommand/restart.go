package subcommand

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/models"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/subscribers"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/events"
)

func NewRestartCommand(soloCtx *context.CliContext) *cobra.Command {
	var profiles []string

	restartCmd := &cobra.Command{
		Use:     "restart",
		GroupID: "lifecycle",
		Short:   "Restarts your app",
		Long:    "Restarts your app",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return soloCtx.Project.ReloadWithProfiles(profiles)
		},
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			model, err := models.NewRestartModel(soloCtx)
			if err != nil {
				return err
			}

			p := tea.NewProgram(model, tea.WithAltScreen())

			eventManager := events.GetEventManagerInstance()
			eventManager.Subscribe(subscribers.NewEventBusToBubbleTeaBridge(soloCtx, p))

			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		}),
	}

	restartCmd.Flags().StringSliceVarP(&profiles, "profile", "", []string{"*"}, "Profiles to use for the command.")

	return restartCmd
}
