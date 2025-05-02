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
	contextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	serviceName, ok := server.Context().Value(contextValueName).(string)
	if !ok {
		t.soloCtx.Logger.Info("Service name not found")
		return fmt.Errorf("unauthorized")
	}

	t.eventManager.Publish(&wms.WorkflowStartedEvent{
		BaseWorkflowEvent: wms.BaseWorkflowEvent{
			ServiceName:  serviceName,
			WorkflowName: workflowName,
		},
	})

	workflow := t.workflowFactory.Make(t.soloCtx.Project, serviceName, workflowName)
	workflowSuccess := true

	if workflow != nil {
		for step := range workflow.StepIterator() {
			err := step.Trigger(func() error {
				// Trigger callback
				t.eventManager.Publish(&wms.WorkflowStepStartedEvent{
					BaseWorkflowEvent: wms.BaseWorkflowEvent{
						ServiceName:  serviceName,
						WorkflowName: workflowName,
					},
					Name: step.GetName(),
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
							ServiceName:  serviceName,
							WorkflowName: workflowName,
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
						ServiceName:  serviceName,
						WorkflowName: workflowName,
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
						ServiceName:  serviceName,
						WorkflowName: workflowName,
					},
					Err: err,
				})

				return err
			}
		}
	}

	t.eventManager.Publish(&wms.WorkflowCompleteEvent{
		BaseWorkflowEvent: wms.BaseWorkflowEvent{
			ServiceName:  serviceName,
			WorkflowName: workflowName,
		},
		Successful: workflowSuccess,
	})

	return server.Send(&services.WorkflowStreamResponse{
		Action: services.WorkflowAction_COMPLETE_ACTION,
	})
}
