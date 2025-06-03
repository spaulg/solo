package subcommand

import (
	"errors"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/entrypoint"
	"github.com/spaulg/solo/internal/pkg/entrypoint/context"
	"github.com/spf13/cobra"
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

			if _, err := commonworkflow.WorkflowNameFromString(args[0]); err != nil {
				return errors.New("unknown event name")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			workflowRunner, err := entrypoint.WorkflowRunnerFactory(entrypointCtx)
			if err != nil {
				panic(err)
			}

			name, _ := commonworkflow.WorkflowNameFromString(args[0])
			if err := workflowRunner.Execute(name); err != nil {
				panic(err)
			}
		},
	}
}
