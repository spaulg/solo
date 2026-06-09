package service_definitions

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/common/infra/grpc/services"
	solo_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
)

type WorkflowServerImpl struct {
	soloCtx *solo_context.CliContext
	services.UnimplementedWorkflowServer
	orchestrator        ContainerImageWorkingDirectoryResolver
	workflowExecTracker WorkflowExecTracker
	workflowRunner      WorkflowRunner
}

func NewWorkflowService(
	soloCtx *solo_context.CliContext,
	orchestrator ContainerImageWorkingDirectoryResolver,
	workflowExecTracker WorkflowExecTracker,
	workflowRunner WorkflowRunner,
) *WorkflowServerImpl {
	return &WorkflowServerImpl{
		soloCtx:             soloCtx,
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
		bidiStreamServer := NewRunWorkflowStreamWrapper(server)
		workflowName := commonworkflow.WorkflowNameFromString(request.RunRequest.WorkflowName)

		return t.handleRunWorkflowRequest(workflowName, bidiStreamServer)
	default:
		return errors.New("unsupported first message")
	}
}

func (t WorkflowServerImpl) handleRunWorkflowRequest(
	workflowName commonworkflow.WorkflowName,
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
) error {

	workflowSession, err := NewWorkflowSession(
		t.soloCtx,
		workflowName,
		server,
		t.workflowExecTracker,
		t.orchestrator,
	)

	if err != nil {
		return fmt.Errorf("failed to create workflow session: %w", err)
	}

	return t.workflowRunner.RunWorkflow(workflowSession)
}
