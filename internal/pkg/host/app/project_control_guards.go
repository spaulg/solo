package app

import (
	"fmt"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/infra/container"
)

func startGuard(
	orchestrator container.Orchestrator,
	containerEntrypointPath string,
) func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
	return func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
		if err := guardCallback(workflowcommon.FirstPreStartContainer); err != nil {
			return err
		}

		if err := guardCallback(workflowcommon.FirstPreStartService); err != nil {
			return err
		}

		if err := guardCallback(workflowcommon.PreStartContainer); err != nil {
			return err
		}

		if err := guardCallback(workflowcommon.PreStartService); err != nil {
			return err
		}

		// Exec post start commands
		postStartEvents := []workflowcommon.WorkflowName{
			workflowcommon.PostStartContainer,
			workflowcommon.PostStartService,
			workflowcommon.FirstPostStartContainer,
			workflowcommon.FirstPostStartService,
		}

		for _, event := range postStartEvents {
			postStartCommand := []string{
				containerEntrypointPath,
				"trigger-event",
				event.String(),
			}

			if err := orchestrator.StartCommand(container, postStartCommand); err != nil {
				return fmt.Errorf("error running compose: %w", err)
			}

			if err := guardCallback(event); err != nil {
				return err
			}
		}

		return nil
	}
}

func stopGuard(
	orchestrator container.Orchestrator,
	containerEntrypointPath string,
) func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
	return func(container string, _ func(_ workflowcommon.WorkflowName) error) error {
		// Exec pre stop commands
		preStopCommand := []string{
			containerEntrypointPath,
			"trigger-event",
			workflowcommon.PreStopContainer.String(),
		}

		if err := orchestrator.StartCommand(container, preStopCommand); err != nil {
			return fmt.Errorf("error running compose: %w", err)
		}

		return nil
	}
}

func destroyGuard(
	orchestrator container.Orchestrator,
	containerEntrypointPath string,
) func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error {
	return func(container string, _ func(_ workflowcommon.WorkflowName) error) error {
		// Exec pre destroy commands
		preDestroyCommand := []string{
			containerEntrypointPath,
			"trigger-event",
			workflowcommon.PreDestroyContainer.String(),
		}

		if err := orchestrator.StartCommand(container, preDestroyCommand); err != nil {
			return fmt.Errorf("error running compose: %w", err)
		}

		return nil
	}
}
