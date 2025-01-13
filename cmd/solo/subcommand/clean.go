package subcommand

import (
	"errors"
	"github.com/spaulg/solo/internal/pkg/solo"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spf13/cobra"
	"os"
)

func NewCleanSubCommand(soloCtx *context.SoloContext) *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean the app",
		Long:  "Clean the app",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return soloCtx.TryLock()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			projectControl, err := solo.ProjectControlFactory(soloCtx)
			if err != nil {
				return err
			}

			return projectControl.Clean(true)
		},
		PostRunE: func(cmd *cobra.Command, args []string) error {
			if err := soloCtx.Unlock(); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return err
				}
			}

			return nil
		},
	}
}
