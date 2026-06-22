package service_definitions

import (
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	commonworkflow "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	services2 "github.com/spaulg/solo/internal/pkg/common/infra/grpc/services"
	"github.com/spaulg/solo/internal/pkg/host/domain"
)

type WorkflowServerImpl struct {
	services2.UnimplementedWorkflowServer

	logger              *slog.Logger
	project             domain.Project
	orchestrator        ContainerImageWorkingDirectoryResolver
	workflowExecTracker WorkflowExecTracker
	workflowRunner      WorkflowRunner
}

func NewWorkflowService(
	logger *slog.Logger,
	project domain.Project,
	orchestrator ContainerImageWorkingDirectoryResolver,
	workflowExecTracker WorkflowExecTracker,
	workflowRunner WorkflowRunner,
) *WorkflowServerImpl {
	return &WorkflowServerImpl{
		logger:              logger,
		project:             project,
		orchestrator:        orchestrator,
		workflowExecTracker: workflowExecTracker,
		workflowRunner:      workflowRunner,
	}
}

func (t WorkflowServerImpl) RunWorkflowStream(
	server grpc.BidiStreamingServer[services2.RunWorkflowStreamRequest, services2.WorkflowStreamResponse],
) error {
	message, err := server.Recv()
	if err != nil {
		return err
	}

	switch request := message.Request.(type) {
	case *services2.RunWorkflowStreamRequest_RunRequest:
		bidiStreamServer := NewRunWorkflowStreamWrapper(server)
		workflowName := commonworkflow.WorkflowNameFromString(request.RunRequest.WorkflowName)

		return t.handleRunWorkflowRequest(workflowName, bidiStreamServer)
	default:
		return errors.New("unsupported first message")
	}
}

func (t WorkflowServerImpl) handleRunWorkflowRequest(
	workflowName commonworkflow.WorkflowName,
	server grpc.BidiStreamingServer[services2.WorkflowStreamRequest, services2.WorkflowStreamResponse],
) error {

	workflowSession, err := NewWorkflowSession(
		t.logger,
		t.project,
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
