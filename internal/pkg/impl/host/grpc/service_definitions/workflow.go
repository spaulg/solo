package service_definitions

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/spaulg/solo/internal/pkg/impl/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	solo_context "github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/grpc/interceptors"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type WorkflowServerImpl struct {
	soloCtx *solo_context.CliContext
	services.UnimplementedWorkflowServer
	eventManager        events_types.Manager
	orchestrator        container_types.Orchestrator
	workflowFactory     wms_types.WorkflowFactory
	workflowExecTracker wms_types.WorkflowExecTracker
}

type serviceContainerDetails struct {
	serviceName       string
	containerName     string
	fullContainerName string
}

func NewWorkflowService(
	soloCtx *solo_context.CliContext,
	eventManager events_types.Manager,
	orchestrator container_types.Orchestrator,
	workflowFactory wms_types.WorkflowFactory,
	workflowExecTracker wms_types.WorkflowExecTracker,
) *WorkflowServerImpl {
	return &WorkflowServerImpl{
		soloCtx:             soloCtx,
		eventManager:        eventManager,
		orchestrator:        orchestrator,
		workflowFactory:     workflowFactory,
		workflowExecTracker: workflowExecTracker,
	}
}

func (t WorkflowServerImpl) RunWorkflowStream(
	server grpc.BidiStreamingServer[services.RunWorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	message, err := server.Recv()
	if err != nil {
		return err
	}

	switch request := message.Request.(type) {
	case *services.RunWorkflowStreamRequest_RunRequest:
		bidiStreamServer := NewRunWorkflowStreamWrapper(server)
		return t.workflowStream(commonworkflow.WorkflowNameFromString(request.RunRequest.WorkflowName), bidiStreamServer)
	default:
		return errors.New("unsupported first message")
	}
}

func (t WorkflowServerImpl) workflowStream(
	workflowName commonworkflow.WorkflowName,
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	// Lookup container/service details
	containerDetails, err := t.lookupContainerDetails(server.Context())
	if err != nil {
		return fmt.Errorf("failed to lookup container details: %w", err)
	}

	// Handle previously run once-only workflows
	hasServiceWorkflowRun, err := t.HasServiceWorkflowRun(containerDetails.serviceName, workflowName)
	if err != nil {
		return fmt.Errorf("failed to check if service workflow has run: %w", err)
	}

	hasFirstContainerWorkflowRun := t.HasFirstContainerWorkflowRun(workflowName, server)

	if hasServiceWorkflowRun || hasFirstContainerWorkflowRun {
		t.eventManager.Publish(&wms_types.WorkflowSkippedEvent{
			BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
				ServiceName:       containerDetails.serviceName,
				ContainerName:     containerDetails.containerName,
				FullContainerName: containerDetails.fullContainerName,
				WorkflowName:      workflowName,
			},
			Successful: true,
		})
	} else {
		workflowSuccess, err := t.applyWorkflowStream(
			workflowName,
			server,
			containerDetails,
		)

		if err != nil {
			return err
		}

		t.eventManager.Publish(&wms_types.WorkflowCompleteEvent{
			BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
				ServiceName:       containerDetails.serviceName,
				ContainerName:     containerDetails.containerName,
				FullContainerName: containerDetails.fullContainerName,
				WorkflowName:      workflowName,
			},
			Successful: workflowSuccess,
		})
	}

	return server.Send(&services.WorkflowStreamResponse{
		Action: services.WorkflowAction_COMPLETE_ACTION,
	})
}

func (t WorkflowServerImpl) applyWorkflowStream(
	workflowName commonworkflow.WorkflowName,
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
	containerDetails *serviceContainerDetails,
) (bool, error) {
	workflowSuccess := true

	t.eventManager.Publish(&wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:       containerDetails.serviceName,
			ContainerName:     containerDetails.containerName,
			FullContainerName: containerDetails.fullContainerName,
			WorkflowName:      workflowName,
		},
	})

	workflow, err := t.workflowFactory.Make(t.soloCtx, t.orchestrator, containerDetails.serviceName, workflowName)
	if err != nil {
		return false, fmt.Errorf("failed to create workflow: %w", err)
	}

	if workflow != nil {
		for step := range workflow.StepIterator() {
			err := step.Trigger(func() error {
				// Trigger callback
				t.eventManager.Publish(&wms_types.WorkflowStepStartedEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:       containerDetails.serviceName,
						ContainerName:     containerDetails.containerName,
						FullContainerName: containerDetails.fullContainerName,
						WorkflowName:      workflowName,
					},
					StepID:    step.GetID(),
					Name:      step.GetName(),
					Command:   step.GetCommand(),
					Arguments: step.GetArguments(),
					Cwd:       step.GetWorkingDirectory(),
					Shell:     step.GetShell(),
				})

				return server.Send(&services.WorkflowStreamResponse{
					Action: services.WorkflowAction_RUN_COMMAND_ACTION,
					RunCommand: &services.WorkflowRunCommand{
						Command:          step.GetCommand(),
						Arguments:        step.GetArguments(),
						WorkingDirectory: step.GetWorkingDirectory(),
					},
				})
			}, func() (*uint8, error) {
				// Progress callback
				result, err := server.Recv()
				if err != nil {
					return nil, err
				}

				if result.Result == services.WorkflowResult_RUN_COMMAND_RESULT {
					var exitCodePtr *uint8
					var exitCode uint8

					if result.RunCommandResult.ExitCode != nil {
						exitCode = uint8(*result.RunCommandResult.ExitCode) // nolint:gosec
						exitCodePtr = &exitCode
					}

					t.eventManager.Publish(&wms_types.WorkflowStepOutputEvent{
						BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
							ServiceName:       containerDetails.serviceName,
							ContainerName:     containerDetails.containerName,
							FullContainerName: containerDetails.fullContainerName,
							WorkflowName:      workflowName,
						},
						StepID: step.GetID(),
						Stdout: result.RunCommandResult.Stdout,
						Stderr: result.RunCommandResult.Stderr,
					})

					return exitCodePtr, nil
				}

				return nil, errors.New("unknown result")
			}, func(exitCode uint8) error {
				// Completion callback
				t.eventManager.Publish(&wms_types.WorkflowStepCompleteEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:       containerDetails.serviceName,
						ContainerName:     containerDetails.containerName,
						FullContainerName: containerDetails.fullContainerName,
						WorkflowName:      workflowName,
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
						ServiceName:       containerDetails.serviceName,
						ContainerName:     containerDetails.containerName,
						FullContainerName: containerDetails.fullContainerName,
						WorkflowName:      workflowName,
					},
					Err: err,
				})

				return workflowSuccess, err
			}

			// If the step failed, skip the remaining steps
			if !workflowSuccess {
				return workflowSuccess, nil
			}
		}
	}

	return workflowSuccess, nil
}

func (t WorkflowServerImpl) HasServiceWorkflowRun(
	serviceName string,
	workflowName commonworkflow.WorkflowName,
) (bool, error) {
	if !workflowName.IsServiceWorkflow() {
		return false, nil
	}

	isFirstExecution, err := t.workflowExecTracker.MarkExecuted(serviceName, workflowName)
	if err != nil {
		return false, fmt.Errorf("failed to mark service workflow as executed: %w", err)
	}

	return !isFirstExecution, nil
}

func (t WorkflowServerImpl) HasFirstContainerWorkflowRun(
	workflowName commonworkflow.WorkflowName,
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) bool {
	if !workflowName.IsFirstContainerWorkflow() {
		return false
	}

	firstWorkflowCompleteContextValueName := interceptors.FirstContainerComplete(workflowName)
	firstWorkflowComplete, firstWorkflowOk := server.Context().Value(firstWorkflowCompleteContextValueName).(string)

	return firstWorkflowOk && firstWorkflowComplete == "true"
}

func (t WorkflowServerImpl) lookupContainerDetails(ctx context.Context) (*serviceContainerDetails, error) {
	// Extract service name
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	serviceName, ok := ctx.Value(serviceNameContextValueName).(string)
	if !ok {
		t.soloCtx.Logger.Error("Service name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	// Extract container name
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	containerName, ok := ctx.Value(containerNameContextValueName).(string)
	if !ok {
		t.soloCtx.Logger.Error("Container name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	// Extract full container name
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	fullContainerName, ok := ctx.Value(fullContainerNameContextValueName).(string)
	if !ok {
		t.soloCtx.Logger.Error("Full container name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	return &serviceContainerDetails{
		serviceName:       serviceName,
		containerName:     containerName,
		fullContainerName: fullContainerName,
	}, nil
}
