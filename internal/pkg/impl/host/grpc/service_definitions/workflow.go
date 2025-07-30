package service_definitions

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/spaulg/solo/internal/pkg/impl/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/grpc/interceptors"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type WorkflowServerImpl struct {
	soloCtx *context.CliContext
	services.UnimplementedWorkflowServer
	eventManager        events_types.Manager
	workflowFactory     wms_types.WorkflowFactory
	workflowExecTracker *WorkflowExecTracker
}

func NewWorkflowService(
	soloCtx *context.CliContext,
	eventManager events_types.Manager,
	workflowFactory wms_types.WorkflowFactory,
	workflowExecTracker *WorkflowExecTracker,
) *WorkflowServerImpl {
	return &WorkflowServerImpl{
		soloCtx:             soloCtx,
		eventManager:        eventManager,
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
	// Extract service name
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	serviceName, ok := server.Context().Value(serviceNameContextValueName).(string)
	if !ok {
		t.soloCtx.Logger.Error("Service name not found")
		return fmt.Errorf("unauthorized")
	}

	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	containerName, ok := server.Context().Value(containerNameContextValueName).(string)
	if !ok {
		t.soloCtx.Logger.Error("Container name not found")
		return fmt.Errorf("unauthorized")
	}

	// todo: handle service level workflow execution skipping
	//		 Service level events need atomic map of service names and worklows that should be skipped.
	//		 This needs to happen here, as it needs to be close to the workflow execution logic.
	//		 There is some overlap with the workflow guard, but this is more about skipping
	//		 the workflow execution, while the guard is about aborting after timeouts.
	//		 The map needs to save/load to disk as a json file.
	//		 When a service is destroyed, its status should be reset so that it does not
	//		 get skipped the next time its started.

	isFirstExecution := false
	var err error

	if workflowName.IsServiceWorkflow() {
		isFirstExecution, err = t.workflowExecTracker.MarkExecuted(serviceName, workflowName)
		if err != nil {
			return fmt.Errorf("failed to mark service workflow as executed: %w", err)
		}
	}

	// First pre start complete
	firstPreStartCompleteContextValueName := interceptors.FirstPreStartComplete(interceptors.FirstPreStartCompleteContextValueName)
	firstPreStartComplete, ok := server.Context().Value(firstPreStartCompleteContextValueName).(string)

	if !isFirstExecution || (workflowName == commonworkflow.FirstPreStartContainer && (ok || firstPreStartComplete == "true")) {
		t.eventManager.Publish(&wms_types.WorkflowSkippedEvent{
			BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
				ServiceName:   serviceName,
				ContainerName: containerName,
				WorkflowName:  workflowName,
			},
			Successful: true,
		})
	} else {
		workflowSuccess, err := t.applyWorkflowStream(workflowName, server, serviceName, containerName)

		if err != nil {
			return err
		}

		t.eventManager.Publish(&wms_types.WorkflowCompleteEvent{
			BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
				ServiceName:   serviceName,
				ContainerName: containerName,
				WorkflowName:  workflowName,
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
	serviceName string,
	containerName string,
) (bool, error) {
	workflowSuccess := true

	t.eventManager.Publish(&wms_types.WorkflowStartedEvent{
		BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
			ServiceName:   serviceName,
			ContainerName: containerName,
			WorkflowName:  workflowName,
		},
	})

	workflow := t.workflowFactory.Make(t.soloCtx.Project, serviceName, workflowName)

	if workflow != nil {
		for step := range workflow.StepIterator() {
			err := step.Trigger(func() error {
				// Trigger callback
				t.eventManager.Publish(&wms_types.WorkflowStepStartedEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:   serviceName,
						ContainerName: containerName,
						WorkflowName:  workflowName,
					},
					StepId:    step.GetId(),
					Name:      step.GetName(),
					Command:   step.GetCommand(),
					Arguments: step.GetArguments(),
					Cwd:       step.GetWorkingDirectory(),
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
						exitCode = uint8(*result.RunCommandResult.ExitCode)
						exitCodePtr = &exitCode
					}

					t.eventManager.Publish(&wms_types.WorkflowStepOutputEvent{
						BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
							ServiceName:   serviceName,
							ContainerName: containerName,
							WorkflowName:  workflowName,
						},
						StepId: step.GetId(),
						Stdout: result.RunCommandResult.Stdout,
						Stderr: result.RunCommandResult.Stderr,
					})

					return exitCodePtr, nil
				} else {
					return nil, errors.New("unknown result")
				}
			}, func(exitCode uint8) error {
				// Completion callback
				t.eventManager.Publish(&wms_types.WorkflowStepCompleteEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:   serviceName,
						ContainerName: containerName,
						WorkflowName:  workflowName,
					},
					StepId:    step.GetId(),
					ExitCode:  exitCode,
					Command:   step.GetCommand(),
					Arguments: step.GetArguments(),
					Cwd:       step.GetWorkingDirectory(),
				})

				if exitCode != 0 {
					workflowSuccess = false
				}

				return nil
			})

			if err != nil {
				t.eventManager.Publish(&wms_types.WorkflowErrorEvent{
					BaseWorkflowEvent: wms_types.BaseWorkflowEvent{
						ServiceName:   serviceName,
						ContainerName: containerName,
						WorkflowName:  workflowName,
					},
					Err: err,
				})

				return workflowSuccess, err
			}
		}
	}

	return workflowSuccess, nil
}
