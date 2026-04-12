package service_definitions

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/common/infra/grpc/services"
	solo_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	wms3 "github.com/spaulg/solo/internal/pkg/impl/host/shared/wms"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
	"github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
)

type WorkflowServerImpl struct {
	soloCtx *solo_context.CliContext
	services.UnimplementedWorkflowServer
	eventManager        events_types.Manager
	orchestrator        container_types.Orchestrator
	workflowExecTracker wms.WorkflowExecTracker
	workflowRunner      wms3.WorkflowRunner
}

func NewWorkflowService(
	soloCtx *solo_context.CliContext,
	eventManager events_types.Manager,
	orchestrator container_types.Orchestrator,
	workflowExecTracker wms.WorkflowExecTracker,
	workflowRunner wms3.WorkflowRunner,
) *WorkflowServerImpl {
	return &WorkflowServerImpl{
		soloCtx:             soloCtx,
		eventManager:        eventManager,
		orchestrator:        orchestrator,
		workflowExecTracker: workflowExecTracker,
		workflowRunner:      workflowRunner,
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
		workflowName := commonworkflow.WorkflowNameFromString(request.RunRequest.WorkflowName)
		bidiStreamServer := NewRunWorkflowStreamWrapper(server)

		workflowSession, err := NewWorkflowSession(
			t.soloCtx,
			workflowName,
			bidiStreamServer,
			t.workflowExecTracker,
			t.orchestrator,
		)

		if err != nil {
			return fmt.Errorf("failed to create workflow session: %w", err)
		}

		return t.handleRunRequest(bidiStreamServer, workflowSession)
	default:
		return errors.New("unsupported first message")
	}
}

func (t WorkflowServerImpl) handleRunRequest(
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
	workflowSession *WorkflowSession,
) error {
	// Handle previously run once-only workflows
	hasServiceWorkflowRun, err := workflowSession.HasServiceWorkflowRun(workflowSession.GetServiceName())
	if err != nil {
		return fmt.Errorf("failed to check if service workflow has run: %w", err)
	}

	hasFirstContainerWorkflowRun := workflowSession.HasFirstContainerWorkflowRun()

	if hasServiceWorkflowRun || hasFirstContainerWorkflowRun {
		t.eventManager.Publish(&wms.WorkflowSkippedEvent{
			BaseWorkflowEvent: wms.BaseWorkflowEvent{
				ServiceName:       workflowSession.GetServiceName(),
				ContainerName:     workflowSession.GetContainerName(),
				FullContainerName: workflowSession.GetFullContainerName(),
				WorkflowName:      workflowSession.GetWorkflowName(),
			},
			Successful: true,
		})
	} else {
		workflowSuccess, err := t.workflowRunner.RunWorkflow(workflowSession)

		if err != nil {
			return err
		}

		t.eventManager.Publish(&wms.WorkflowCompleteEvent{
			BaseWorkflowEvent: wms.BaseWorkflowEvent{
				ServiceName:       workflowSession.GetServiceName(),
				ContainerName:     workflowSession.GetContainerName(),
				FullContainerName: workflowSession.GetFullContainerName(),
				WorkflowName:      workflowSession.GetWorkflowName(),
			},
			Successful: workflowSuccess,
		})
	}

	return server.Send(&services.WorkflowStreamResponse{
		Action: services.WorkflowAction_COMPLETE_ACTION,
	})
}
