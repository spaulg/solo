package service_definitions

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
	"google.golang.org/grpc"
)

type WorkflowServerImpl struct {
	soloCtx *context.SoloContext
	services.UnimplementedWorkflowServer
	eventManager    events.Manager
	workflowFactory wms.Factory
}

func NewWorkflowService(
	soloCtx *context.SoloContext,
	eventManager events.Manager,
	workflowFactory wms.Factory,
) *WorkflowServerImpl {
	return &WorkflowServerImpl{
		soloCtx:         soloCtx,
		eventManager:    eventManager,
		workflowFactory: workflowFactory,
	}
}

func (t WorkflowServerImpl) BuildWorkflowStream(
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	return t.workflowStream(commonworkflow.Build, server)
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

func (t WorkflowServerImpl) workflowStream(
	workflowName commonworkflow.Name,
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {
	// Extract service name
	serviceName, ok := server.Context().Value("ServiceName").(string)
	if !ok {
		t.soloCtx.Logger.Info("Service name not found")
		return fmt.Errorf("unauthorized")
	}

	t.eventManager.Publish(&wms.WorkflowStartedEvent{
		BaseEvent: events.BaseEvent{
			ServiceName:  serviceName,
			WorkflowName: workflowName,
		},
	})

	workflow := t.workflowFactory.Make(t.soloCtx.Project, serviceName, workflowName)

	for step := range workflow.StepIterator() {
		err := step.Trigger(func() error {
			// Trigger callback
			t.eventManager.Publish(&wms.WorkflowStepStartedEvent{
				BaseEvent: events.BaseEvent{
					ServiceName:  serviceName,
					WorkflowName: workflowName,
				},
				Name: step.GetName(),
			})

			return server.Send(&services.WorkflowStreamResponse{
				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
				RunCommand: &services.WorkflowRunCommand{
					Command:          step.GetCommand(),
					Arguments:        step.GetCommandArguments(),
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
					BaseEvent: events.BaseEvent{
						ServiceName:  serviceName,
						WorkflowName: workflowName,
					},
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
				BaseEvent: events.BaseEvent{
					ServiceName:  serviceName,
					WorkflowName: workflowName,
				},
				ExitCode: exitCode,
			})

			return nil
		})

		if err != nil {
			t.eventManager.Publish(&wms.WorkflowErrorEvent{
				BaseEvent: events.BaseEvent{
					ServiceName:  serviceName,
					WorkflowName: workflowName,
				},
				Err: err,
			})

			return err
		}
	}

	t.eventManager.Publish(&wms.WorkflowCompleteEvent{
		BaseEvent: events.BaseEvent{
			ServiceName:  serviceName,
			WorkflowName: workflowName,
		},
	})

	return server.Send(&services.WorkflowStreamResponse{
		Action: services.WorkflowAction_COMPLETE_ACTION,
	})
}
