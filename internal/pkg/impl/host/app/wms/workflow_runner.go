package wms

import (
	"fmt"

	solo_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	wms_shared "github.com/spaulg/solo/internal/pkg/impl/host/shared/wms"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
)

type WorkflowRunner struct {
	soloCtx         *solo_context.CliContext
	eventManager    events_types.Manager
	workflowFactory wms_types.WorkflowFactory
}

func NewWorkflowRunner(
	soloCtx *solo_context.CliContext,
	eventManager events_types.Manager,
	workflowFactory wms_types.WorkflowFactory,
) *WorkflowRunner {
	return &WorkflowRunner{
		soloCtx:         soloCtx,
		eventManager:    eventManager,
		workflowFactory: workflowFactory,
	}
}

func (t *WorkflowRunner) RunWorkflow(workflowSession wms_shared.WorkflowSession) (bool, error) {
	workflowSuccess := true

	t.eventManager.Publish(&wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       workflowSession.GetServiceName(),
			ContainerName:     workflowSession.GetContainerName(),
			FullContainerName: workflowSession.GetFullContainerName(),
			WorkflowName:      workflowSession.GetWorkflowName(),
		},
	})

	serviceWorkingDirectory, err := workflowSession.GetWorkingDirectory()
	if err != nil {
		return false, err
	}

	workflow, err := t.workflowFactory.Make(
		t.soloCtx,
		workflowSession.GetServiceName(),
		serviceWorkingDirectory,
		workflowSession.GetWorkflowName(),
	)

	if err != nil {
		return false, fmt.Errorf("failed to create workflow: %w", err)
	}

	if workflow != nil {
		for step := range workflow.StepIterator() {
			err := step.Trigger(func() error {
				// Trigger callback
				t.eventManager.Publish(&wms_types.WorkflowStepStartedEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:       workflowSession.GetServiceName(),
						ContainerName:     workflowSession.GetContainerName(),
						FullContainerName: workflowSession.GetFullContainerName(),
						WorkflowName:      workflowSession.GetWorkflowName(),
					},
					StepID:    step.GetID(),
					Name:      step.GetName(),
					Command:   step.GetCommand(),
					Arguments: step.GetArguments(),
					Cwd:       step.GetWorkingDirectory(),
					Shell:     step.GetShell(),
				})

				return workflowSession.RunCommand(&wms_shared.RunCommandRequest{
					Command:          step.GetCommand(),
					Arguments:        step.GetArguments(),
					WorkingDirectory: step.GetWorkingDirectory(),
				})
			}, func() (*uint8, error) {
				// Progress callback
				result, err := workflowSession.RecvCommandResponse()
				if err != nil {
					return nil, err
				}

				t.eventManager.Publish(&wms_types.WorkflowStepOutputEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:       workflowSession.GetServiceName(),
						ContainerName:     workflowSession.GetContainerName(),
						FullContainerName: workflowSession.GetFullContainerName(),
						WorkflowName:      workflowSession.GetWorkflowName(),
					},
					StepID: step.GetID(),
					Stdout: result.Stdout,
					Stderr: result.Stderr,
				})

				return result.ExitCode, nil
			}, func(exitCode uint8) error {
				// Completion callback
				t.eventManager.Publish(&wms_types.WorkflowStepCompleteEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:       workflowSession.GetServiceName(),
						ContainerName:     workflowSession.GetContainerName(),
						FullContainerName: workflowSession.GetFullContainerName(),
						WorkflowName:      workflowSession.GetWorkflowName(),
					},
					StepID:    step.GetID(),
					ExitCode:  exitCode,
					Command:   step.GetCommand(),
					Arguments: step.GetArguments(),
					Cwd:       step.GetWorkingDirectory(),
					Shell:     step.GetShell(),
				})

				if exitCode != 0 {
					workflowSuccess = false
				}

				return nil
			})

			if err != nil {
				t.eventManager.Publish(&wms_types.WorkflowErrorEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:       workflowSession.GetServiceName(),
						ContainerName:     workflowSession.GetContainerName(),
						FullContainerName: workflowSession.GetFullContainerName(),
						WorkflowName:      workflowSession.GetWorkflowName(),
					},
					Err: err,
				})

				return false, err
			}

			// If the step failed, skip the remaining steps
			if !workflowSuccess {
				return workflowSuccess, nil
			}
		}
	}

	return workflowSuccess, nil
}
