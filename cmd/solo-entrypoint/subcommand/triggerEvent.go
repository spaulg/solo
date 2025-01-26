package subcommand

import (
	"errors"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/entrypoint"
	"github.com/spf13/cobra"
)

func NewTriggerEventCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "trigger-event [event]",
		Short: "Trigger a provisioning event",
		Long:  "Trigger a provisioning event",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires exactly one argument")
			}

			if _, err := commonworkflow.FromString(args[0]); err != nil {
				return errors.New("unknown event name")
			}

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			workflowRunner, err := entrypoint.WorkflowRunnerFactory()
			if err != nil {
				panic(err)
			}

			name, _ := commonworkflow.FromString(args[0])
			workflowRunner.Execute(name)
		},
	}
}
