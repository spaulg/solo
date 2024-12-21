package service_definitions

import (
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
	"google.golang.org/grpc"
)

type WorkflowServerImpl struct {
	soloCtx *context.SoloContext
	services.UnimplementedWorkflowServer
	eventManager    events.Manager
	project         *project.Project
	workflowFactory wms.Factory
}

func NewWorkflowService(
	soloCtx *context.SoloContext,
	eventManager events.Manager,
	project *project.Project,
	workflowFactory wms.Factory,
) *WorkflowServerImpl {
	return &WorkflowServerImpl{
		soloCtx:         soloCtx,
		eventManager:    eventManager,
		project:         project,
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

	workflow := t.workflowFactory.Make(t.project, serviceName, workflowName)

	for step := range workflow.StepIterator() {
		err := step.Trigger(func() error {
			// Trigger callback
			return server.Send(&services.WorkflowStreamResponse{
				Action: services.WorkflowAction_RUN_COMMAND_ACTION,
				RunCommand: &services.WorkflowRunCommand{
					Command:          step.GetCommand(),
					Arguments:        step.GetCommandArguments(),
					WorkingDirectory: step.GetWorkingDirectory(),
				},
			})
		}, func() (*wms.StepProgress, error) {
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

				// Notify progress subscribers
				t.eventManager.Publish(events.CommandProgress, &events.Event{
					ServiceName:  serviceName,
					WorkflowName: workflowName,
					// todo: add data
				})

				return &wms.StepProgress{
					ExitCode: exitCodePtr,
					Stdout:   &result.RunCommandResult.Stdout,
					Stderr:   &result.RunCommandResult.Stderr,
				}, nil
			} else {
				return nil, errors.New("unknown result")
			}
		}, func() error {
			// Completion callback
			t.eventManager.Publish(events.CommandFinished, &events.Event{
				ServiceName:  serviceName,
				WorkflowName: workflowName,
				// todo: add data
			})

			return nil
		})

		if err != nil {
			return err
		}
	}

	t.eventManager.Publish(events.WorkflowFinished, &events.Event{
		ServiceName:  serviceName,
		WorkflowName: workflowName,
		// todo: add data
	})

	t.soloCtx.Logger.Error("Workflow finished")
	return server.Send(&services.WorkflowStreamResponse{
		Action: services.WorkflowAction_COMPLETE_ACTION,
	})
}
