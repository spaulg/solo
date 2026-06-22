package wms

import (
	"fmt"

	"github.com/spaulg/solo/internal/pkg/host/app/event_manager/events"
	wf2 "github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type WorkflowRunner struct {
	config          *domain2.Config
	project         domain2.Project
	eventManager    events.Manager
	workflowFactory wf2.Factory
}

func NewWorkflowRunner(
	config *domain2.Config,
	project domain2.Project,
	eventManager events.Manager,
	workflowFactory wf2.Factory,
) *WorkflowRunner {
	return &WorkflowRunner{
		config:          config,
		project:         project,
		eventManager:    eventManager,
		workflowFactory: workflowFactory,
	}
}

func (t *WorkflowRunner) RunWorkflow(workflowSession wf2.Session) error {
	// Handle previously run once-only workflows
	hasServiceWorkflowRun, err := workflowSession.HasServiceWorkflowRun(workflowSession.GetServiceName())
	if err != nil {
		return fmt.Errorf("failed to check if service workflow has run: %w", err)
	}

	hasFirstContainerWorkflowRun := workflowSession.HasFirstContainerWorkflowRun()

	if hasServiceWorkflowRun || hasFirstContainerWorkflowRun {
		t.eventManager.Publish(&wf2.SkippedEvent{
			BaseWorkflowEvent: wf2.BaseWorkflowEvent{
				ServiceName:       workflowSession.GetServiceName(),
				ContainerName:     workflowSession.GetContainerName(),
				FullContainerName: workflowSession.GetFullContainerName(),
				WorkflowName:      workflowSession.GetWorkflowName(),
			},
			Successful: true,
		})
	} else {
		workflowSuccess, err := t.handleRunWorkflow(workflowSession)

		if err != nil {
			return err
		}

		t.eventManager.Publish(&wf2.CompleteEvent{
			BaseWorkflowEvent: wf2.BaseWorkflowEvent{
				ServiceName:       workflowSession.GetServiceName(),
				ContainerName:     workflowSession.GetContainerName(),
				FullContainerName: workflowSession.GetFullContainerName(),
				WorkflowName:      workflowSession.GetWorkflowName(),
			},
			Successful: workflowSuccess,
		})
	}

	return workflowSession.MarkCompletion()
}

func (t *WorkflowRunner) handleRunWorkflow(workflowSession wf2.Session) (bool, error) {
	workflowSuccess := true

	t.eventManager.Publish(&wf2.StartedEvent{
		BaseWorkflowEvent: wf2.BaseWorkflowEvent{
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
		t.config,
		t.project,
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
				t.eventManager.Publish(&wf2.StepStartedEvent{
					BaseWorkflowEvent: wf2.BaseWorkflowEvent{
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

				return workflowSession.RunCommand(&wf2.RunCommandRequest{
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

				t.eventManager.Publish(&wf2.StepOutputEvent{
					BaseWorkflowEvent: wf2.BaseWorkflowEvent{
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
				t.eventManager.Publish(&wf2.StepCompleteEvent{
					BaseWorkflowEvent: wf2.BaseWorkflowEvent{
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
				t.eventManager.Publish(&wf2.ErrorEvent{
					BaseWorkflowEvent: wf2.BaseWorkflowEvent{
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
