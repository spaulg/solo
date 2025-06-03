package service_definitions

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/interceptors"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
	"google.golang.org/grpc"
)

type WorkflowServerImpl struct {
	soloCtx *context.CliContext
	services.UnimplementedWorkflowServer
	eventManager    events.Manager
	workflowFactory wms.Factory
}

func NewWorkflowService(
	soloCtx *context.CliContext,
	eventManager events.Manager,
	workflowFactory wms.Factory,
) *WorkflowServerImpl {
	return &WorkflowServerImpl{
		soloCtx:         soloCtx,
		eventManager:    eventManager,
		workflowFactory: workflowFactory,
	}
}

func (t WorkflowServerImpl) FirstPreStartWorkflowStream(
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	return t.workflowStream(commonworkflow.FirstPreStart, server)
}

func (t WorkflowServerImpl) PreStartWorkflowStream(
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	return t.workflowStream(commonworkflow.PreStart, server)
}

func (t WorkflowServerImpl) PostStartWorkflowStream(
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	return t.workflowStream(commonworkflow.PostStart, server)
}

func (t WorkflowServerImpl) PreStopWorkflowStream(
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	return t.workflowStream(commonworkflow.PreStop, server)
}

func (t WorkflowServerImpl) PreDestroyWorkflowStream(
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	return t.workflowStream(commonworkflow.PreDestroy, server)
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

	// First pre start complete
	firstPreStartCompleteContextValueName := interceptors.FirstPreStartComplete(interceptors.FirstPreStartCompleteContextValueName)
	firstPreStartComplete, ok := server.Context().Value(firstPreStartCompleteContextValueName).(string)

	if workflowName == commonworkflow.FirstPreStart && (ok || firstPreStartComplete == "true") {
		t.eventManager.Publish(&wms.WorkflowSkippedEvent{
			BaseWorkflowEvent: wms.BaseWorkflowEvent{
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

		t.eventManager.Publish(&wms.WorkflowCompleteEvent{
			BaseWorkflowEvent: wms.BaseWorkflowEvent{
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

	t.eventManager.Publish(&wms.WorkflowStartedEvent{
		BaseWorkflowEvent: wms.BaseWorkflowEvent{
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
				t.eventManager.Publish(&wms.WorkflowStepStartedEvent{
					BaseWorkflowEvent: wms.BaseWorkflowEvent{
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

					t.eventManager.Publish(&wms.WorkflowStepOutputEvent{
						BaseWorkflowEvent: wms.BaseWorkflowEvent{
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
				t.eventManager.Publish(&wms.WorkflowStepCompleteEvent{
					BaseWorkflowEvent: wms.BaseWorkflowEvent{
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
				t.eventManager.Publish(&wms.WorkflowErrorEvent{
					BaseWorkflowEvent: wms.BaseWorkflowEvent{
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
