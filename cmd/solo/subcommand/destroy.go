package subcommand

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/models"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/subscribers"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/events"
)

func NewDestroySubCommand(soloCtx *context.CliContext) *cobra.Command {
	var destroyCmdYes bool
	var profiles []string

	destroyCmd := &cobra.Command{
		Use:     "destroy",
		GroupID: "lifecycle",
		Short:   "Destroys your app",
		Long:    "Destroys your app",
		Annotations: map[string]string{
			RequireConfigLoadSuccessAnnotation:  "true",
			RequireProjectLoadSuccessAnnotation: "true",
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := soloCtx.Project.ReloadWithProfiles(profiles); err != nil {
				return err
			}

			if !destroyCmdYes {
				var cmdConfirmationString string
				for {
					fmt.Print("Are you sure you want to destroy (y/n)? ")
					_, err := fmt.Scanln(&cmdConfirmationString)

					if err != nil {
						continue
					} else if strings.ToLower(cmdConfirmationString) == "n" {
						os.Exit(0)
					} else if strings.ToLower(cmdConfirmationString) == "y" {
						break
					}
				}
			}

			return nil
		},
		RunE: soloCtx.ProtectWithLock(func(cmd *cobra.Command, args []string) error {
			model, err := models.NewDestroyModel(soloCtx, profiles)
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

	destroyCmd.Flags().BoolVarP(&destroyCmdYes, "yes", "y", false, "Answer yes non-interactively to confirmation questions")
	destroyCmd.Flags().StringSliceVarP(&profiles, "profile", "", []string{"*"}, "Profiles to use for the command.")

	return destroyCmd
}
