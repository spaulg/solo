package subcommand

import (
	"errors"

	"github.com/spf13/cobra"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/app"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/app/context"
)

func NewTriggerEventCommand(entrypointCtx *context.EntrypointContext) *cobra.Command {
	return &cobra.Command{
		Use:   "trigger-event [event]",
		Short: "Trigger a provisioning event",
		Long:  "Trigger a provisioning event",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires exactly one argument")
			}

			if commonworkflow.Undefined == commonworkflow.WorkflowNameFromString(args[0]) {
				return errors.New("unknown event name")
			}

			return nil
		},
		Run: func(_ *cobra.Command, args []string) {
			workflowRunner, err := app.WorkflowRunnerFactory(entrypointCtx)
			if err != nil {
				panic(err)
			}

			name := commonworkflow.WorkflowNameFromString(args[0])
			if err := workflowRunner.Execute(name); err != nil {
				panic(err)
			}
		},
	}
}
