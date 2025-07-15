package subcommand

import (
	"errors"

	"github.com/spf13/cobra"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/context"
)

func NewTriggerEventCommand(entrypointCtx *context.EntrypointContext) *cobra.Command {
	return &cobra.Command{
		Use:   "trigger-event [event]",
		Short: "Trigger a provisioning event",
		Long:  "Trigger a provisioning event",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires exactly one argument")
			}

			if commonworkflow.Undefined == commonworkflow.WorkflowNameFromString(args[0]) {
				return errors.New("unknown event name")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			workflowRunner, err := entrypoint.WorkflowRunnerFactory(entrypointCtx)
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
