package service_definitions

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/impl/common/infra/grpc/services"
	solo_context "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc/interceptors"
	wms_shared "github.com/spaulg/solo/internal/pkg/impl/host/shared/wms"
	"github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	"github.com/spaulg/solo/internal/pkg/types/host/infra/container"
)

type WorkflowSession struct {
	soloCtx             *solo_context.CliContext
	workflowName        commonworkflow.WorkflowName
	server              grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse]
	workflowExecTracker wms.WorkflowExecTracker
	orchestrator        container.Orchestrator

	serviceName       string
	containerName     string
	fullContainerName string
}

func NewWorkflowSession(
	soloCtx *solo_context.CliContext,
	workflowName commonworkflow.WorkflowName,
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
	workflowExecTracker wms.WorkflowExecTracker,
	orchestrator container.Orchestrator,
) (*WorkflowSession, error) {
	ctx := server.Context()

	// Extract service name
	serviceNameContextValueName := interceptors.ServiceName(interceptors.ServiceNameContextValueName)
	serviceName, ok := ctx.Value(serviceNameContextValueName).(string)
	if !ok {
		soloCtx.Logger.Error("Service name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	// Extract container name
	containerNameContextValueName := interceptors.ContainerName(interceptors.ContainerNameContextValueName)
	containerName, ok := ctx.Value(containerNameContextValueName).(string)
	if !ok {
		soloCtx.Logger.Error("Container name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	// Extract full container name
	fullContainerNameContextValueName := interceptors.ContainerName(interceptors.FullContainerNameContextValueName)
	fullContainerName, ok := ctx.Value(fullContainerNameContextValueName).(string)
	if !ok {
		soloCtx.Logger.Error("Full container name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	return &WorkflowSession{
		soloCtx:             soloCtx,
		workflowName:        workflowName,
		server:              server,
		workflowExecTracker: workflowExecTracker,
		orchestrator:        orchestrator,
		serviceName:         serviceName,
		containerName:       containerName,
		fullContainerName:   fullContainerName,
	}, nil
}

func (t *WorkflowSession) GetWorkflowName() commonworkflow.WorkflowName {
	return t.workflowName
}

func (t *WorkflowSession) HasServiceWorkflowRun(
	serviceName string,
) (bool, error) {
	if !t.workflowName.IsServiceWorkflow() {
		return false, nil
	}

	isFirstExecution, err := t.workflowExecTracker.MarkExecuted(serviceName, t.workflowName)
	if err != nil {
		return false, fmt.Errorf("failed to mark service workflow as executed: %w", err)
	}

	return !isFirstExecution, nil
}

func (t *WorkflowSession) HasFirstContainerWorkflowRun() bool {
	if !t.workflowName.IsFirstContainerWorkflow() {
		return false
	}

	firstWorkflowCompleteContextValueName := interceptors.FirstContainerComplete(t.workflowName)
	firstWorkflowComplete, firstWorkflowOk := t.server.Context().Value(firstWorkflowCompleteContextValueName).(string)

	return firstWorkflowOk && firstWorkflowComplete == "true"
}

func (t *WorkflowSession) GetServiceName() string {
	return t.serviceName
}

func (t *WorkflowSession) GetContainerName() string {
	return t.containerName
}

func (t *WorkflowSession) GetFullContainerName() string {
	return t.fullContainerName
}

func (t *WorkflowSession) GetWorkingDirectory() (string, error) {
	var err error

	serviceName := t.GetServiceName()
	service := t.soloCtx.Project.Services().GetService(serviceName)
	serviceWorkingDirectory := service.GetConfig().WorkingDir

	if serviceWorkingDirectory == "" {
		// Project does not define a working directory for the service
		// Use orchestrator to lookup the working directory from the service image
		serviceWorkingDirectory, err = t.orchestrator.ResolveImageWorkingDirectory(serviceName)
		if err != nil {
			return "", err
		}
	}

	return serviceWorkingDirectory, err
}

func (t *WorkflowSession) RunCommand(request *wms_shared.RunCommandRequest) error {
	return t.server.Send(&services.WorkflowStreamResponse{
		Action: services.WorkflowAction_RUN_COMMAND_ACTION,
		RunCommand: &services.WorkflowRunCommand{
			Command:          request.Command,
			Arguments:        request.Arguments,
			WorkingDirectory: request.WorkingDirectory,
		},
	})
}

func (t *WorkflowSession) RecvCommandResponse() (*wms_shared.CommandResponse, error) {
	result, err := t.server.Recv()

	if err != nil {
		return nil, err
	}

	if result.Result != services.WorkflowResult_RUN_COMMAND_RESULT {
		return nil, errors.New("unknown result")
	}

	var exitCodePtr *uint8
	var exitCode uint8

	if result.RunCommandResult.ExitCode != nil {
		exitCode = uint8(*result.RunCommandResult.ExitCode) // nolint:gosec
		exitCodePtr = &exitCode
	}

	return &wms_shared.CommandResponse{
		Stdout:   result.RunCommandResult.Stdout,
		Stderr:   result.RunCommandResult.Stderr,
		ExitCode: exitCodePtr,
	}, nil
}
