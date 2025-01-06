package subcommand

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spaulg/solo/internal/pkg/solo/bubbletea/models"
	"github.com/spaulg/solo/internal/pkg/solo/bubbletea/subscribers"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func NewRebuildCommand(soloCtx *context.CliContext) *cobra.Command {
	var rebuildCmdYes bool

	rebuildCmd := &cobra.Command{
		Use:     "rebuild",
		GroupID: "lifecycle",
		Short:   "Rebuilds your app from scratch, preserving data",
		Long:    "Rebuilds your app from scratch, preserving data",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if !rebuildCmdYes {
				var cmdConfirmationString string
				for {
					fmt.Print("Are you sure you want to rebuild (y/n)? ")
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
			model, err := models.NewRebuildModel(soloCtx)
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

	rebuildCmd.Flags().BoolVarP(&rebuildCmdYes, "yes", "y", false, "Answer yes non-interactively to confirmation questions")

	return rebuildCmd
}
