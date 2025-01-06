package subcommand

import (
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/models"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/subscribers"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/events"
)

func NewStopCommand(soloCtx *context.CliContext) *cobra.Command {
	var profiles []string

	stopCmd := &cobra.Command{
		Use:     "stop",
		GroupID: "lifecycle",
		Short:   "Stops your app",
		Long:    "Stops your app",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return soloCtx.Project.ReloadWithProfiles(profiles)
		},
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			model, err := models.NewStopModel(soloCtx)
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

	stopCmd.Flags().StringSliceVarP(&profiles, "profile", "", []string{"*"}, "Profiles to use for the command.")

	return stopCmd
}
